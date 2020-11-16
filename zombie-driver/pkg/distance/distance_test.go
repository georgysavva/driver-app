package distance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/heetch/georgysavva-technical-test/zombie-driver/pkg/distance"
)

func TestCalculate(t *testing.T) {
	t.Parallel()
	actual := distance.Calculate(48.864193, 2.350498, 48.863193, 2.351498)
	assert.Equal(t, 133.24684781984962, actual)
}
