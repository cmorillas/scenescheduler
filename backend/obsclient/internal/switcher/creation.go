// backend/obsclient/internal/switcher/creation.go
//
// This file contains methods for creating OBS inputs and scene items.
// It handles the initial staging of new sources in the temporary scene.
//
// Contents:
// - Input and Scene Item Creation
// - Settings Merging

package switcher

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"scenescheduler/backend/eventbus"
)

// ============================================================================
// INPUT AND SCENE ITEM CREATION
// ============================================================================

// inputExists checks if an input with the given name exists
func (s *Switcher) inputExists(client *goobs.Client, inputName string) (bool, error) {
	_, err := client.Inputs.GetInputSettings(&inputs.GetInputSettingsParams{
		InputName: &inputName,
	})
	if err != nil {
		// If the error is about the input not existing, return false
		if err.Error() == "specified input does not exist" {
			return false, nil
		}
		return false, fmt.Errorf("error checking if input exists: %w", err)
	}
	return true, nil
}

// removeInputIfExists removes an input if it exists, ignoring errors if it doesn't
func (s *Switcher) removeInputIfExists(client *goobs.Client, inputName string) error {
	exists, err := s.inputExists(client, inputName)
	if err != nil {
		return fmt.Errorf("error checking if input exists for removal: %w", err)
	}

	if exists {
		s.logger.Debug("Removing existing input", "inputName", inputName)
		_, err = client.Inputs.RemoveInput(&inputs.RemoveInputParams{
			InputName: &inputName,
		})
		if err != nil {
			return fmt.Errorf("failed to remove existing input: %w", err)
		}
	}
	return nil
}

// createOBSInput prepares a new source and its scene item in the temporary scene.
func (s *Switcher) createOBSInput(client *goobs.Client, program *eventbus.Program) (int, error) {
	tmpScene := s.config.ScheduleSceneAux
	prefixedName := s.config.SourceNamePrefix + program.SourceName

	// Validate that the input kind is supported by this instance of OBS.
	listResp, err := client.Inputs.GetInputKindList()
	if err != nil {
		return 0, fmt.Errorf("could not fetch supported input kinds: %w", err)
	}

	if !slices.Contains(listResp.InputKinds, program.InputKind) {
		s.logger.Error("Input kind not supported by OBS",
			"requested", program.InputKind,
			"supported", listResp.InputKinds)
		return 0, fmt.Errorf("input kind %q is not supported by OBS", program.InputKind)
	}

	// Remove existing input if it exists
	if err := s.removeInputIfExists(client, prefixedName); err != nil {
		s.logger.Warn("Failed to remove existing input, will attempt to continue", 
			"error", err,
			"inputName", prefixedName)
	}

	// Create the input and its scene item in the temporary "staging" scene, initially hidden.
	respCreate, err := s.createInputInScene(client, tmpScene, prefixedName, program, false)
	if err != nil {
		return 0, fmt.Errorf("failed to create input in temp scene: %w", err)
	}

	s.logger.Debug("Successfully created scene item in temp scene",
		"sceneItemId", respCreate.SceneItemId,
		"scene", tmpScene)
	return respCreate.SceneItemId, nil
}

// createInputInScene is a helper that handles the raw creation of the input and its scene item.
func (s *Switcher) createInputInScene(client *goobs.Client, sceneName, inputName string, program *eventbus.Program, enabled bool) (*inputs.CreateInputResponse, error) {
	respDefaults, err := client.Inputs.GetInputDefaultSettings(&inputs.GetInputDefaultSettingsParams{
		InputKind: &program.InputKind,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get default settings for kind '%s': %w", program.InputKind, err)
	}
	baseSettings := respDefaults.DefaultInputSettings

	parsedURL, err := url.Parse(program.URI)
	isURL := err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""

	switch program.InputKind {
	case "ffmpeg_source":
		if isURL {
			baseSettings["input"] = program.URI
			baseSettings["is_local_file"] = false
		} else {
			baseSettings["local_file"] = program.URI
			baseSettings["is_local_file"] = true
		}
	case "vlc_source":
		baseSettings["playlist"] = []map[string]interface{}{{"value": program.URI}}
	case "browser_source":
		baseSettings["url"] = program.URI
	}

	finalSettings := s.mergeSettings(baseSettings, program.InputSettings)
	params := &inputs.CreateInputParams{
		SceneName:        &sceneName,
		InputName:        &inputName,
		InputKind:        &program.InputKind,
		InputSettings:    finalSettings,
		SceneItemEnabled: &[]bool{false}[0], // Always create hidden
	}

	s.logger.Debug("Attempting to create OBS input", "inputName", inputName, "scene", sceneName)
	return client.Inputs.CreateInput(params)
}

// ============================================================================
// SETTINGS MERGING
// ============================================================================

// mergeSettings merges custom settings from the user over a base settings map.
func (s *Switcher) mergeSettings(base map[string]interface{}, custom interface{}) map[string]interface{} {
	if custom == nil {
		return base
	}
	customMap, ok := custom.(map[string]interface{})
	if !ok {
		s.logger.Warn("Could not merge settings: program.InputSettings is not a valid map", "settings", custom)
		return base
	}

	for key, value := range customMap {
		base[key] = value
	}
	return base
}