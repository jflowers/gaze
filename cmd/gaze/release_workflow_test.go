package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseWorkflow_MacOSSigningIdentity(t *testing.T) {
	workflowPath := filepath.Join("..", "..", ".github", "workflows", "release.yml")
	workflow, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	contents := string(workflow)
	checkStart := strings.Index(contents, "  check-signing-secrets:")
	if checkStart < 0 {
		t.Fatal("release workflow must contain check-signing-secrets job")
	}
	checkEnd := strings.Index(contents[checkStart:], "\n  push-unsigned-cask:")
	if checkEnd < 0 {
		t.Fatal("release workflow must contain push-unsigned-cask after check-signing-secrets")
	}
	checkSigningSecrets := contents[checkStart : checkStart+checkEnd]

	for _, name := range []string{
		"MACOS_SIGN_P12",
		"MACOS_SIGN_PASSWORD",
		"MACOS_SIGN_IDENTITY",
		"MACOS_NOTARY_KEY",
		"MACOS_NOTARY_KEY_ID",
		"MACOS_NOTARY_ISSUER_ID",
	} {
		mapping := name + ": ${{ secrets." + name + " }}"
		if !strings.Contains(checkSigningSecrets, mapping) {
			t.Errorf("check-signing-secrets must contain %q", mapping)
		}

		warning := `echo "::warning::Signing configuration missing: ` + name + `"`
		if !strings.Contains(checkSigningSecrets, warning) {
			t.Errorf("check-signing-secrets must contain %q", warning)
		}
	}

	for _, line := range strings.Split(checkSigningSecrets, "\n") {
		if strings.Contains(line, "::warning::") && strings.Contains(line, "$") {
			t.Errorf("signing-secret warning must not interpolate values: %q", line)
		}
	}

	signMacOSStart := strings.Index(contents, "  sign-macos:")
	if signMacOSStart < 0 {
		t.Fatal("release workflow must contain sign-macos job")
	}
	signMacOS := contents[signMacOSStart:]
	identityMapping := `MACOS_SIGN_IDENTITY: ${{ secrets.MACOS_SIGN_IDENTITY }}`
	if !strings.Contains(signMacOS, identityMapping) {
		t.Errorf("sign-macos must contain %q", identityMapping)
	}

	for _, expected := range []string{
		`if [ -n "$MACOS_SIGN_P12" ] && [ -n "$MACOS_SIGN_PASSWORD" ] && [ -n "$MACOS_SIGN_IDENTITY" ] && [ -n "$MACOS_NOTARY_KEY" ] && [ -n "$MACOS_NOTARY_KEY_ID" ] && [ -n "$MACOS_NOTARY_ISSUER_ID" ]; then`,
		`needs.check-signing-secrets.outputs.has_signing_secrets == 'true'`,
		`needs.check-signing-secrets.outputs.has_signing_secrets == 'false'`,
	} {
		if !strings.Contains(contents, expected) {
			t.Errorf("release workflow must contain %q", expected)
		}
	}

	for _, forbidden := range []string{
		`vars.MACOS_SIGN_IDENTITY`,
		`Developer ID Application:`,
	} {
		if strings.Contains(contents, forbidden) {
			t.Errorf("release workflow must not contain %q", forbidden)
		}
	}
}
