/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

// np-release creates the update signing key and prepares GitHub release
// assets for NET PORTAL.
//
//	go run ./cmd/np-release keygen [-key PATH]
//	    Creates a secret signing key (keep it private, never commit it) and
//	    prints the public key to put into updater/constants.go.
//
//	go run ./cmd/np-release sign [-key PATH] [-out DIR]
//	    After build.bat, copies amd64/x86/arm64 netportal.exe into DIR as
//	    netportal-<arch>-<version>.exe and writes the signed latest.sig.
//	    Upload everything in DIR to a GitHub release tagged v<version>.
//
// The signature format is the one `signify -S -e` produces, which is what
// the updater verifies.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"

	"github.com/amnezia-vpn/amneziawg-windows-client/version"
)

const (
	secretKeyComment = "untrusted comment: NET PORTAL release secret key — KEEP PRIVATE"
	sigComment       = "untrusted comment: NET PORTAL release, verify with the key in updater/constants.go"
)

func defaultKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "netportal-release.key"
	}
	return filepath.Join(home, ".netportal", "release.key")
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "keygen":
		fs := flag.NewFlagSet("keygen", flag.ExitOnError)
		key := fs.String("key", defaultKeyPath(), "where to write the secret key")
		fs.Parse(os.Args[2:])
		err = keygen(*key)
	case "sign":
		fs := flag.NewFlagSet("sign", flag.ExitOnError)
		key := fs.String("key", defaultKeyPath(), "secret key created by keygen")
		out := fs.String("out", filepath.Join("release", "v"+version.Number), "output directory")
		fs.Parse(os.Args[2:])
		err = sign(*key, *out)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: np-release keygen [-key PATH] | sign [-key PATH] [-out DIR]")
	os.Exit(2)
}

func keygen(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists; refusing to overwrite a signing key", path)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	keyNum := make([]byte, 8)
	if _, err := rand.Read(keyNum); err != nil {
		return err
	}
	secret := append(append([]byte("Ed"), keyNum...), priv...)
	public := append(append([]byte("Ed"), keyNum...), pub...)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	content := secretKeyComment + "\n" + base64.StdEncoding.EncodeToString(secret) + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return err
	}
	fmt.Printf("Secret key written to %s\n", path)
	fmt.Println("Back it up somewhere safe. Anyone with it can push updates to your users;")
	fmt.Println("if you lose it, installed copies can no longer be updated.")
	fmt.Println()
	fmt.Println("Put this into updater/constants.go as ReleasePublicKey:")
	fmt.Println(base64.StdEncoding.EncodeToString(public))
	return nil
}

func readSecretKey(path string) (keyNum []byte, priv ed25519.PrivateKey, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[len(lines)-1]))
	if err != nil || len(raw) != 2+8+ed25519.PrivateKeySize || raw[0] != 'E' || raw[1] != 'd' {
		return nil, nil, errors.New("not a NET PORTAL secret key")
	}
	return raw[2:10], ed25519.PrivateKey(raw[10:]), nil
}

func copyAndHash(src, dst string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	h, _ := blake2b.New256(nil)
	if _, err := io.Copy(io.MultiWriter(out, h), in); err != nil {
		out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sign(keyPath, outDir string) error {
	keyNum, priv, err := readSecretKey(keyPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	ver := version.Number
	var lines []string
	for _, arch := range []string{"amd64", "x86", "arm64"} {
		src := filepath.Join(arch, "netportal.exe")
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("%s not found — run build.bat first", src)
		}
		if !bytes.Contains(mustRead(src), []byte(ver)) {
			return fmt.Errorf("%s does not look like version %s — rebuild after changing version/version.go", src, ver)
		}
		name := fmt.Sprintf("netportal-%s-%s.exe", arch, ver)
		sum, err := copyAndHash(src, filepath.Join(outDir, name))
		if err != nil {
			return err
		}
		lines = append(lines, sum+"  "+name)
		fmt.Printf("  %s\n", name)
	}
	sort.Strings(lines)
	message := []byte(strings.Join(lines, "\n") + "\n")
	sig := ed25519.Sign(priv, message)
	full := append(append([]byte("Ed"), keyNum...), sig...)
	content := sigComment + "\n" + base64.StdEncoding.EncodeToString(full) + "\n" + string(message)
	if err := os.WriteFile(filepath.Join(outDir, "latest.sig"), []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("  latest.sig\n\nRelease files are in %s.\n", outDir)
	fmt.Printf("Create a GitHub release with the tag v%s and upload all of them.\n", ver)
	return nil
}

func mustRead(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return b
}
