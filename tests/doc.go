// Package tests builds a test site with the theme and checks the output.
//
// Run it from the theme repo:
//
//	go -C tests test -count=1 ./...
//
// It needs Hugo Extended. The check that runs main.js in a browser also needs
// Chromium, and is skipped without it. Pass -count=1, since Go doesn't see
// changes to the theme or to Hugo and would reuse an old result.
package tests
