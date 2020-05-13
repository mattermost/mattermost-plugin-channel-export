package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/assert"
)

func TestIsLicensed(t *testing.T) {
	t.Run("no license", func(t *testing.T) {
		assert.False(t, isLicensed(nil))
	})

	t.Run("nil features", func(t *testing.T) {
		assert.False(t, isLicensed(&model.License{}))
	})

	t.Run("nil future features", func(t *testing.T) {
		assert.False(t, isLicensed(&model.License{Features: &model.Features{}}))
	})

	t.Run("disabled future features", func(t *testing.T) {
		falseValue := false
		assert.False(t, isLicensed(&model.License{Features: &model.Features{
			FutureFeatures: &falseValue,
		}}))
	})

	t.Run("enabled future features", func(t *testing.T) {
		trueValue := true
		assert.True(t, isLicensed(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}))
	})
}
