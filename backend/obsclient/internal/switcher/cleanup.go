// backend/obsclient/internal/switcher/cleanup.go
//
// This file contains methods for cleaning up and removing OBS resources.
// It handles both targeted cleanup (specific programs) and failsafe cleanup (orphans).
//
// Contents:
// - Specific Program Cleanup
// - Orphaned Resources Cleanup
// - Rollback Removal

package switcher

import (
	"fmt"
	"strings"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// SPECIFIC PROGRAM CLEANUP
// ============================================================================

// cleanupSpecificProgram removes a known program from the scene.
// This is used to clean up the previous program after a successful switch.
// Returns error only if the operation fails unexpectedly (not if resource is already gone).
func (s *Switcher) cleanupSpecificProgram(client *goobs.Client, sceneName string, program *eventbus.Program) error {
	prefixedName := s.config.SourceNamePrefix + program.SourceName

	s.logger.Debug("Cleaning up specific program", "name", prefixedName, "scene", sceneName)

	// Try to get the scene item ID
	idResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
		SceneName:  &sceneName,
		SourceName: &prefixedName,
	})

	// If item exists, hide it first then remove
	if err == nil {
		// CLEANUP: Hide is best-effort
		_, _ = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName:        &sceneName,
			SceneItemId:      &idResp.SceneItemId,
			SceneItemEnabled: &[]bool{false}[0],
		})

		// CLEANUP: Remove is best-effort
		_, _ = client.SceneItems.RemoveSceneItem(&sceneitems.RemoveSceneItemParams{
			SceneName:   &sceneName,
			SceneItemId: &idResp.SceneItemId,
		})
	} else {
		s.logger.Debug("Scene item not found (may have been removed already)",
			"name", prefixedName,
			"error", err)
	}

	// Always try to remove the input source (idempotent)
	// CLEANUP: Input removal is best-effort
	if _, err := client.Inputs.RemoveInput(&inputs.RemoveInputParams{
		InputName: &prefixedName,
	}); err != nil {
		s.logger.Debug("Could not remove input source (may not exist)",
			"name", prefixedName,
			"error", err)
	}

	return nil
}

// ============================================================================
// ORPHANED RESOURCES CLEANUP
// ============================================================================

// cleanupOrphanedManagedSources removes managed sources that aren't current or target.
// This is a failsafe to catch resources from failed previous switches.
// Returns error only for unexpected failures, not for missing resources.
func (s *Switcher) cleanupOrphanedManagedSources(
	client *goobs.Client,
	sceneName string,
	current, target *eventbus.Program,
) error {
	prefix := s.config.SourceNamePrefix

	// Build set of protected source names
	protectedSources := make(map[string]bool)
	if current != nil {
		protectedSources[prefix+current.SourceName] = true
	}
	if target != nil {
		protectedSources[prefix+target.SourceName] = true
	}

	resp, err := client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{
		SceneName: &sceneName,
	})
	if err != nil {
		return fmt.Errorf("could not get scene item list: %w", err)
	}

	orphanCount := 0
	for _, item := range resp.SceneItems {
		// If this is a managed source (has our prefix) and it's NOT protected
		if strings.HasPrefix(item.SourceName, prefix) && !protectedSources[item.SourceName] {
			orphanCount++
			s.logger.InfoGui("Removing orphaned managed source", "name", item.SourceName)

			// CLEANUP: Hide and remove are best-effort
			_, _ = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
				SceneName:        &sceneName,
				SceneItemId:      &item.SceneItemID,
				SceneItemEnabled: &[]bool{false}[0],
			})
			_, _ = client.SceneItems.RemoveSceneItem(&sceneitems.RemoveSceneItemParams{
				SceneName:   &sceneName,
				SceneItemId: &item.SceneItemID,
			})
		}
	}

	if orphanCount > 0 {
		s.logger.InfoGui("Cleaned up orphaned sources", "count", orphanCount)
	}

	return nil
}

// ============================================================================
// ROLLBACK REMOVAL
// ============================================================================

// removeOBSInput handles the idempotent removal of a scene item and its underlying input source.
// This is used for rollback scenarios and is best-effort (does not return errors).
func (s *Switcher) removeOBSInput(client *goobs.Client, sceneName string, program *eventbus.Program) error {
	prefixedName := s.config.SourceNamePrefix + program.SourceName

	// CLEANUP: Best-effort removal
	idResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
		SceneName:  &sceneName,
		SourceName: &prefixedName,
	})
	if err == nil {
		_, _ = client.SceneItems.RemoveSceneItem(&sceneitems.RemoveSceneItemParams{
			SceneName:   &sceneName,
			SceneItemId: &idResp.SceneItemId,
		})
	}

	_, _ = client.Inputs.RemoveInput(&inputs.RemoveInputParams{
		InputName: &prefixedName,
	})

	return nil
}