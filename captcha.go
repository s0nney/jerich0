// Copyright 2011 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package captcha implements generation and verification of image and audio
// CAPTCHAs.
//
// A captcha solution is the sequence of digits 0-9 with the defined length.
// There are two captcha representations: image and audio.
//
// An image representation is a PNG-encoded image with the solution printed on
// it in such a way that makes it hard for computers to solve it using OCR.
//
// An audio representation is a WAVE-encoded (8 kHz unsigned 8-bit) sound with
// the spoken solution (currently in English, Russian, Chinese, and Japanese).
// To make it hard for computers to solve audio captcha, the voice that
// pronounces numbers has random speed and pitch, and there is a randomly
// generated background noise mixed into the sound.
//
// This package doesn't require external files or libraries to generate captcha
// representations; it is self-contained.
//
// To make captchas one-time, the package includes a memory storage that stores
// captcha ids, their solutions, and expiration time. Used captchas are removed
// from the store immediately after calling Verify or VerifyString, while
// unused captchas (user loaded a page with captcha, but didn't submit the
// form) are collected automatically after the predefined expiration time.
// Developers can also provide custom store (for example, which saves captcha
// ids and solutions in database) by implementing Store interface and
// registering the object with SetCustomStore.
//
// Captchas are created by calling New, which returns the captcha id.  Their
// representations, though, are created on-the-fly by calling WriteImage or
// WriteAudio functions. Created representations are not stored anywhere, but
// subsequent calls to these functions with the same id will write the same
// captcha solution. Reload function will create a new different solution for
// the provided captcha, allowing users to "reload" captcha if they can't solve
// the displayed one without reloading the whole page.  Verify and VerifyString
// are used to verify that the given solution is the right one for the given
// captcha id.
//
// Server provides an http.Handler which can serve image and audio
// representations of captchas automatically from the URL. It can also be used
// to reload captchas.  Refer to Server function documentation for details, or
// take a look at the example in "capexample" subdirectory.
package captcha

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	// Default number of digits in captcha solution.
	DefaultLen = 6
	// The number of captchas created that triggers garbage collection used
	// by default store.
	CollectNum = 100
	// Expiration time of captchas used by default store.
	Expiration = 10 * time.Minute
)

var (
	ErrNotFound = errors.New("captcha: id not found")
	// globalStore is a shared storage for captchas, generated by New function.
	globalStore = NewMemoryStore(CollectNum, Expiration)
)

// SetCustomStore sets custom storage for captchas, replacing the default
// memory store. This function must be called before generating any captchas.
func SetCustomStore(s Store) {
	globalStore = s
}

// New creates a new captcha with the standard length, saves it in the internal
// storage and returns its id.
func New() string {
	return NewLen(DefaultLen)
}

// NewLen is just like New, but accepts length of a captcha solution as the
// argument.
func NewLen(length int) (id string) {
	id = randomId()
	// Store the indices (0-35) not the characters
	digits := randomBytesMod(length, 36)
	globalStore.Set(id, digits)
	return
}

// Reload generates and remembers new digits for the given captcha id.  This
// function returns false if there is no captcha with the given id.
//
// After calling this function, the image or audio presented to a user must be
// refreshed to show the new captcha representation (WriteImage and WriteAudio
// will write the new one).
func Reload(id string) bool {
	old := globalStore.Get(id, false)
	if old == nil {
		return false
	}
	globalStore.Set(id, RandomDigits(len(old)))
	return true
}

// WriteImage writes PNG-encoded image representation of the captcha with the
// given id. The image will have the given width and height.
func WriteImage(w io.Writer, id string, width, height int) error {
	d := globalStore.Get(id, false)
	if d == nil {
		return ErrNotFound
	}
	_, err := NewImage(id, d, width, height).WriteTo(w)
	return err
}

// Verify returns true if the given digits are the ones that were used to
// create the given captcha id.
//
// The function deletes the captcha with the given id from the internal
// storage, so that the same captcha can't be verified anymore.

func Verify(id string, digits []byte) bool {
	if digits == nil || len(digits) == 0 {
		return false
	}
	reald := globalStore.Get(id, true)
	if reald == nil {
		return false
	}

	// Temporary debug logging
	fmt.Printf("Stored: %v\nInput: %v\n", reald, digits)

	return bytes.Equal(digits, reald)
}

func VerifyString(id string, answer string) bool {
	// Convert answer to indices (0-35)
	ns := make([]byte, 0, len(answer))
	for _, c := range answer {
		switch {
		case '0' <= c && c <= '9':
			ns = append(ns, byte(c-'0')) // 0-9
		case 'A' <= c && c <= 'Z':
			ns = append(ns, byte(c-'A'+10)) // A=10, B=11, etc.
		case 'a' <= c && c <= 'z':
			ns = append(ns, byte(c-'a'+10)) // lowercase
		default:
			return false
		}
	}
	return Verify(id, ns)
}
