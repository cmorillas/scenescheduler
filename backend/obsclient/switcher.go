// backend/obsclient/switcher.go
//
// This file contains the high-level logic related to scene switching and setup.
// It acts as a bridge between the OBSClient's FSM/event handlers and the
// specialized, internal switcher component.
//
// Contents:
// - State Convergence Logic
// - Program Switching Logic
// - Scene Setup Logic
// - Scene Setup Helpers

package obsclient

import (
	"fmt"
	"scenescheduler/backend/eventbus"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
)

// ============================================================================
// STATE CONVERGENCE LOGIC
// ============================================================================

// convergeToState is the core logic for acting on a TargetProgramState event.
// It compares the desired target with the client's current active program
// and triggers a switch only if they are different.
//
// This method uses a dedicated switchMu to serialize all convergence operations,
// preventing race conditions when multiple TargetProgramState events arrive
// in quick succession.
func (c *OBSClient) convergeToState(state eventbus.TargetProgramState) {
	// CRITICAL: Acquire switch lock FIRST to serialize convergence operations.
	// This prevents race conditions where multiple events could trigger
	// simultaneous switches, leading to unpredictable final state.
	c.switchMu.Lock()
	defer c.switchMu.Unlock()

	// Step 1: Check if convergence is needed (with read lock)
	c.stateMu.RLock()
	needsSwitch := !isProgramSame(c.activeProgram, state.TargetProgram)
	currentProgram := c.activeProgram
	c.stateMu.RUnlock()

	// Idempotency check: If the target is already active, do nothing.
	if !needsSwitch {
		return
	}

	c.logger.Debug("State divergence detected, initiating program switch.",
		"current", getProgramTitle(currentProgram),
		"target", getProgramTitle(state.TargetProgram),
	)

	targetProgram := state.TargetProgram

	// Step 2: Perform the switch without holding stateMu.
	// The switchMu ensures only one switch happens at a time.
	// This allows other goroutines to read state while switch is in progress.
	err := c.performProgramSwitch(targetProgram, state.SeekOffset)
	if err != nil {
		c.logger.Error("Program switch failed", "error", err)
		// On failure, we do not update the active program, maintaining the last known good state.
		return
	}

	// Step 3: Update internal state (with write lock)
	c.stateMu.Lock()
	c.activeProgram = targetProgram
	c.stateMu.Unlock()

	c.logger.Info("Successfully switched program.", "newActiveProgram", getProgramTitle(targetProgram))
}

// isProgramSame compares two ProgramData objects to see if they represent the
// same program. It handles nil pointers gracefully.
func isProgramSame(a, b *eventbus.Program) bool {
	if a == nil && b == nil {
		return true // Both are "no program"
	}
	if a == nil || b == nil {
		return false // One is a program, the other is not
	}
	return a.ID == b.ID // The unique ID is the source of truth
}

// ============================================================================
// PROGRAM SWITCHING LOGIC
// ============================================================================

// performProgramSwitch orchestrates the program switching action by delegating
// to the internal switcher component.
func (c *OBSClient) performProgramSwitch(target *eventbus.Program, offset time.Duration) error {
	if c.connection == nil || c.connection.client == nil {
		return ErrNotConnected
	}

	// Get current program to pass as context
	c.stateMu.RLock()
	current := c.activeProgram
	c.stateMu.RUnlock()

	// Pass both current and target to switcher
    result, err := c.switcher.PerformSwitch(c.connection.client, current, target, offset)
    if err != nil {
        return fmt.Errorf("internal switcher failed: %w", err)
    }

    // Publish the event ONLY if there was a real change
    if result.CurrentProgram != nil || result.PreviousProgram != nil {
        event := eventbus.OBSProgramChanged{
            Timestamp:       result.Timestamp,
            PreviousProgram: result.PreviousProgram,
            CurrentProgram:  result.CurrentProgram,
            SeekOffsetMs:    result.SeekOffsetMs,
        }
        eventbus.Publish(c.bus, event)
        c.logger.Debug("Published OBSProgramChanged event",
            "previous", getProgramTitle(result.PreviousProgram),
            "current", getProgramTitle(result.CurrentProgram))
    }

    return nil
}

// ============================================================================
// SCENE SETUP LOGIC
// ============================================================================

// setupScene is the implementation for the scene setup action, executed
// once the client connects to OBS. It ensures both scenes exist and clears
// the temporary staging scene. The main scene is left untouched to avoid
// visual artifacts - convergence will handle it on the next evaluation cycle.
func (c *OBSClient) setupScene() error {
	client, _ := c.getActiveClientAndContext()
	if client == nil {
		return ErrNotConnected
	}
	mainScene := c.config.ScheduleScene
	auxScene := c.config.ScheduleSceneAux

	c.logger.Debug("Starting scene setup", "mainScene", mainScene, "auxScene", auxScene)

	// 1. Ensure both scenes exist.
	if err := c.ensureSceneExists(client, mainScene); err != nil {
		return fmt.Errorf("failed to ensure main scene '%s' exists: %w", mainScene, err)
	}
	if err := c.ensureSceneExists(client, auxScene); err != nil {
		return fmt.Errorf("failed to ensure aux scene '%s' exists: %w", auxScene, err)
	}

	// 2. Clear the temporary scene of any remnants.
	// This ensures a clean staging area for future switches.
	if err := c.clearAllSceneItems(client, auxScene); err != nil {
		return fmt.Errorf("failed to cleanup aux scene %q: %w", auxScene, err)
	}

	// 3. Also clean the main scene on setup.
	// Rationale: This establishes a known-good, clean state upon connection.
	// The client's internal state `activeProgram` is nil, and now the scene
	// in OBS will also be empty. This prevents an immediate, false "divergence"
	// when the first state event is received from the scheduler.
	if err := c.clearAllSceneItems(client, mainScene); err != nil {
		return fmt.Errorf("failed to cleanup main scene %q: %w", mainScene, err)
	}

	c.logger.Debug("Scene setup completed. Both scenes have been cleared.")

	return nil
}

// ============================================================================
// SCENE SETUP HELPERS
// ============================================================================

// ensureSceneExists checks if a scene exists, creating it if necessary.
func (c *OBSClient) ensureSceneExists(client *goobs.Client, sceneName string) error {
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return fmt.Errorf("could not get scene list: %w", err)
	}
	for _, scene := range resp.Scenes {
		if scene.SceneName == sceneName {
			return nil
		}
	}

	c.logger.Debug("Required scene not found, creating it.", "sceneName", sceneName)
	_, err = client.Scenes.CreateScene(&scenes.CreateSceneParams{
		SceneName: &sceneName,
	})
	return err
}

// clearAllSceneItems removes all items from a given scene.
func (c *OBSClient) clearAllSceneItems(client *goobs.Client, sceneName string) error {
	resp, err := client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{
		SceneName: &sceneName,
	})
	if err != nil {
		return fmt.Errorf("could not get item list for scene '%s': %w", sceneName, err)
	}
	if len(resp.SceneItems) == 0 {
		return nil
	}

	c.logger.Debug("Clearing all items from scene.", "sceneName", sceneName, "itemCount", len(resp.SceneItems))
	for _, item := range resp.SceneItems {
		_, _ = client.SceneItems.RemoveSceneItem(&sceneitems.RemoveSceneItemParams{
			SceneName:   &sceneName,
			SceneItemId: &item.SceneItemID,
		})
	}
	return nil
}


