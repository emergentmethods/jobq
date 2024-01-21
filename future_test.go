package jobq

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuture_SetAndGetResult(t *testing.T) {
	future := NewFuture()
	expectedResult := "test result"
	go future.SetResult(expectedResult, nil)

	result, err := future.Result()
	assert.Nil(t, err, "Future should not return an error")
	assert.Equal(t, expectedResult, result, "Future should return the correct result")
}

func TestFuture_ErrorResult(t *testing.T) {
	future := NewFuture()
	expectedError := errors.New("test error")
	go future.SetResult(nil, expectedError)

	_, err := future.Result()
	assert.Equal(t, expectedError, err, "Future should return the correct error")
}
