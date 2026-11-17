package domain_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/puriice/waypoint/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveStringJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(domain.NewSensitiveString("secret"))
	require.NoError(t, err)
	assert.JSONEq(t, `"secret"`, string(data))

	var got domain.SensitiveString
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, "secret", got.Expose())
}

func TestSensitiveStringUnmarshalRejectsNonString(t *testing.T) {
	var got domain.SensitiveString
	require.Error(t, json.Unmarshal([]byte(`42`), &got))
}

func TestSensitiveStringStaysRedactedInFormatting(t *testing.T) {
	assert.Equal(t, "[REDACTED]", fmt.Sprint(domain.NewSensitiveString("secret")))
}
