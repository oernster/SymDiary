package main

import (
	"strings"
	"testing"
)

// wantedAddress is where the donate button must send a browser, written out
// here rather than read from product.DonateURL.
//
// That is a deliberate second statement of the address, not a lapse in DRY: a
// test comparing the constant against itself asserts nothing, so a typo in the
// payment address would ship green and send a supporter to a page that is not
// Oliver's. One of the two is the value, the other is the claim about it; they
// are only allowed to agree.
const wantedAddress = "https://www.paypal.com/ncp/payment/4XP3AYNMPQGUC"

// recordingOpener stands in for the desktop, keeping every address it was
// handed so a test never opens a real browser.
type recordingOpener struct{ asked []string }

// Open keeps the address instead of opening it.
func (o *recordingOpener) Open(address string) { o.asked = append(o.asked, address) }

func TestDonateAsksTheDesktopForTheDonationPage(t *testing.T) {
	t.Parallel()
	app, _, _ := window(t)
	opener := &recordingOpener{}
	app.opener = opener

	if err := app.Donate(); err != nil {
		t.Fatalf("Donate = %v, want no error", err)
	}
	if len(opener.asked) != 1 {
		t.Fatalf("the desktop was asked for %d addresses, want exactly 1: %v",
			len(opener.asked), opener.asked)
	}
	if opener.asked[0] != wantedAddress {
		t.Errorf("Donate opened %q, want %q", opener.asked[0], wantedAddress)
	}
}

func TestTheDonationAddressIsSecure(t *testing.T) {
	t.Parallel()
	// A payment page reached over plain HTTP is one a supporter should not be
	// sent to, whatever else is right about it.
	if !strings.HasPrefix(wantedAddress, "https://") {
		t.Errorf("the donation address is %q, which is not https", wantedAddress)
	}
}

func TestDonateOpensNothingWhenTheDesktopIsMissing(t *testing.T) {
	t.Parallel()
	// A facade built without an opener is a wiring mistake rather than a state
	// the user can reach, so it must answer the same internal fault every other
	// bound method answers, not end the window.
	app, _, _ := window(t)
	app.opener = nil
	if err := app.Donate(); err == nil {
		t.Error("Donate with no opener answered no error, want the internal fault")
	}
}
