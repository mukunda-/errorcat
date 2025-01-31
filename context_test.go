package errorcat_test

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mukunda.com/errorcat"
	cat "go.mukunda.com/errorcat"
)

func TestDuplicateRecoverError(t *testing.T) {

	assert.NotPanics(t, func() {
		ct := errorcat.NewContext(nil)
		defer errorcat.Recover(ct, "test")
	})

	assert.Panics(t, func() {
		ct := errorcat.NewContext(nil)
		defer errorcat.Recover(ct, "test")
		defer errorcat.Recover(ct, "test")
	})
}

func TestRecoverFinalizer(t *testing.T) {
	// Smoke test for now.
	func() {
		errorcat.NewContext(nil)
		runtime.GC()
	}()
}

func TestGoWithContext(t *testing.T) {
	err := <-cat.Go(func(ct cat.Context) error {
		ct.Catch(true, "whoops")
		return nil
	}, "test")

	assert.Error(t, err)
	assert.Equal(t, "test: whoops", err.Error())
}

func TestForgotDefer(t *testing.T) {
	func() {
		ct := errorcat.NewContext(nil)
		errorcat.Recover(ct, "test")
		assert.Panics(t, func() {
			ct.Catch(true, "whoops")
		})
	}()

}
