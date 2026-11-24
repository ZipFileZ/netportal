/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package updater

// NET PORTAL updates come from the releases of a GitHub repository.
//
// A release tagged "v<version>" must contain:
//   - netportal-<arch>-<version>.exe for amd64, x86 and arm64
//   - latest.sig, the signed list of their BLAKE2b-256 hashes
//
// Both are produced by `go run ./cmd/np-release sign <version>`. Updates
// stay disabled until UpdateRepo and ReleasePublicKey are set.
const (
	// GitHub "owner/repo", for example "someone/netportal". Must be public.
	UpdateRepo = "ZipFileZ/netportal"

	// Public key printed by `go run ./cmd/np-release keygen` (signify format).
	ReleasePublicKey = "RWQoItJM2rzeXSkoVZxF5ju/aEDoayZuxyOxzMyYT2Vl0eBFvIYeeP8W"

	updateServerHost     = "github.com"
	updateServerPort     = 443
	updateServerUseHttps = true
	latestVersionPath    = "/" + UpdateRepo + "/releases/latest/download/latest.sig"
	assetPath            = "/" + UpdateRepo + "/releases/download/v%s/%s"
	assetArchPrefix      = "netportal-%s-"
	assetSuffix          = ".exe"
)

// Enabled reports whether this build is configured to update itself.
func Enabled() bool {
	return UpdateRepo != "" && ReleasePublicKey != ""
}
