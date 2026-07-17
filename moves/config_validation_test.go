package moves

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
)

type configuredDealForValidation struct {
	DealCountComponents
}

type configuredDefaultForValidation struct {
	Default
}

func (*configuredDefaultForValidation) Apply(boardgame.State) error { return nil }

func TestValidateCustomConfigurationRejectsUnusedOption(t *testing.T) {
	err := validateCustomConfiguration(new(configuredDefaultForValidation), boardgame.PropertyCollection{
		configPropTargetCount: 3,
		configPropAmount:      2,
	})
	if err == nil {
		t.Fatal("unused specialized configuration was accepted")
	}
	for _, want := range []string{"WithAmount", "WithTargetCount"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q did not name %s", err, want)
		}
	}
}

func TestValidateCustomConfigurationAcceptsPromotedConsumer(t *testing.T) {
	err := validateCustomConfiguration(new(configuredDealForValidation), boardgame.PropertyCollection{
		configPropTargetCount:    3,
		configPropNumRounds:      2,
		configPropGameProperty:   "DrawStack",
		configPropPlayerProperty: "Hand",
	})
	if err != nil {
		t.Fatalf("valid promoted consumers were rejected: %v", err)
	}
}

func TestValidateCustomConfigurationRejectsHelperOnlyOption(t *testing.T) {
	err := validateCustomConfiguration(new(configuredDefaultForValidation), boardgame.PropertyCollection{configPropManualStart: true})
	if err == nil || !strings.Contains(err.Error(), "WithManualStart") {
		t.Fatalf("WithManualStart misuse error = %v", err)
	}
}

func TestAutoConfigurerRejectsUnusedOption(t *testing.T) {
	var configErr error
	_, _ = newGameManager(func(manager *boardgame.GameManager) []boardgame.MoveConfig {
		auto := NewAutoConfigurer(manager.Delegate())
		_, configErr = auto.Config(new(moveCurrentPlayerDraw), WithTargetCount(3))
		return nil
	})
	if configErr == nil || !strings.Contains(configErr.Error(), "WithTargetCount") || !strings.Contains(configErr.Error(), "has no effect") {
		t.Fatalf("AutoConfigurer.Config error = %v, want unused WithTargetCount diagnostic", configErr)
	}
}
