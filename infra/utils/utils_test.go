package utils_test

import (
	"encoder/infra/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsJson(t *testing.T) {
	json := `
{
	"id": "",
	"file_path" : "sample.mp4"	 
}`

	err := utils.IsJson(json)

	require.Nil(t, err)

	json = "wes"

	err = utils.IsJson(json)
	require.Error(t, err)
}
