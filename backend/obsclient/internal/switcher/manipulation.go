// backend/obsclient/internal/switcher/manipulation.go
//
// This file contains methods for manipulating scene items: applying transforms,
// duplicating between scenes, and controlling visibility.
//
// Contents:
// - Transform Application
// - Scene Item Duplication
// - Visibility Control

package switcher

import (
	"encoding/json"
	"fmt"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/typedefs"
)

// ============================================================================
// TRANSFORM APPLICATION
// ============================================================================

// applyTransformsToSceneItem applies the default and user-defined transforms.
func (s *Switcher) applyTransformsToSceneItem(client *goobs.Client, sceneName string, sceneItemID int, userTransformData interface{}) error {
	if err := s.applyDefaultTransform(client, sceneName, sceneItemID); err != nil {
		s.logger.Warn("Failed to apply default transform to scene item.", "error", err)
	}
	finalTransform, err := s.mergeUserTransform(client, sceneName, sceneItemID, userTransformData)
	if err != nil {
		return fmt.Errorf("could not merge user transform for item %d: %w", sceneItemID, err)
	}
	_, err = client.SceneItems.SetSceneItemTransform(&sceneitems.SetSceneItemTransformParams{
		SceneName:          &sceneName,
		SceneItemId:        &sceneItemID,
		SceneItemTransform: finalTransform,
	})
	return err
}

// applyDefaultTransform sets a scene item's transform to stretch to the canvas bounds.
func (s *Switcher) applyDefaultTransform(client *goobs.Client, sceneName string, sceneItemID int) error {
	videoSettings, err := client.Config.GetVideoSettings()
	if err != nil {
		return fmt.Errorf("could not get canvas video settings: %w", err)
	}
	transform := &typedefs.SceneItemTransform{
		BoundsType:   "OBS_BOUNDS_STRETCH",
		Alignment:    5, // Top-Left
		PositionX:    0,
		PositionY:    0,
		BoundsWidth:  float64(videoSettings.BaseWidth),
		BoundsHeight: float64(videoSettings.BaseHeight),
	}
	_, err = client.SceneItems.SetSceneItemTransform(&sceneitems.SetSceneItemTransformParams{
		SceneName:          &sceneName,
		SceneItemId:        &sceneItemID,
		SceneItemTransform: transform,
	})
	return err
}

// mergeUserTransform merges a user-defined transform object over the current item's transform.
func (s *Switcher) mergeUserTransform(client *goobs.Client, sceneName string, sceneItemID int, userTransformData interface{}) (*typedefs.SceneItemTransform, error) {
	resp, err := client.SceneItems.GetSceneItemTransform(&sceneitems.GetSceneItemTransformParams{
		SceneName:   &sceneName,
		SceneItemId: &sceneItemID,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get current transform to merge: %w", err)
	}
	finalTransform := resp.SceneItemTransform
	if userTransformData == nil {
		return finalTransform, nil
	}
	jsonBytes, err := json.Marshal(userTransformData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user transform data: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, finalTransform); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user transform over base: %w", err)
	}
	return finalTransform, nil
}

// ============================================================================
// SCENE ITEM DUPLICATION
// ============================================================================

// duplicateSceneItem copies a scene item from a source to a destination scene.
func (s *Switcher) duplicateSceneItem(client *goobs.Client, fromScene, toScene string, sceneItemID int) (int, error) {
	s.logger.Debug("Duplicating scene item", "from", fromScene, "to", toScene, "id", sceneItemID)
	resp, err := client.SceneItems.DuplicateSceneItem(&sceneitems.DuplicateSceneItemParams{
		SceneName:            &fromScene,
		DestinationSceneName: &toScene,
		SceneItemId:          &sceneItemID,
	})
	if err != nil {
		return 0, fmt.Errorf("API call to duplicate scene item %d failed: %w", sceneItemID, err)
	}
	return resp.SceneItemId, nil
}

// ============================================================================
// VISIBILITY CONTROL
// ============================================================================

// setSceneItemEnabled changes the visibility of a scene item.
func (s *Switcher) setSceneItemEnabled(client *goobs.Client, sceneName string, sceneItemID int, enabled bool) error {
	s.logger.Debug("Setting scene item visibility", "scene", sceneName, "id", sceneItemID, "enabled", enabled)
	_, err := client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
		SceneName:        &sceneName,
		SceneItemId:      &sceneItemID,
		SceneItemEnabled: &enabled,
	})
	return err
}