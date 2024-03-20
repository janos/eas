// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eas_test

import (
	"context"
	"testing"

	"resenje.org/eas"
)

func TestEASContract_Attest(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	schemaUID := registerSchema(t, client, "string message")

	_, wait, err := client.EAS.Attest(ctx, schemaUID, &eas.AttestOptions{Revocable: true}, "Hello!")
	assertNilError(t, err)

	client.backend.Commit()

	r, err := wait(ctx)
	assertNilError(t, err)
	assertEqual(t, "schema uid", r.Schema, schemaUID)
	assertEqual(t, "attester", r.Attester, client.account)
}

func TestEASContract_GetAttestation(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	schemaUID := registerSchema(t, client, "string message")

	_, wait, err := client.EAS.Attest(ctx, schemaUID, &eas.AttestOptions{Revocable: true}, "Hello!")
	assertNilError(t, err)

	client.backend.Commit()

	r, err := wait(ctx)
	assertNilError(t, err)

	a, err := client.EAS.GetAttestation(ctx, r.UID)
	assertNilError(t, err)

	assertEqual(t, "schema uid", a.Schema, schemaUID)
	assertEqual(t, "attester", a.Attester, client.account)
}

func TestEASContract_GetAttestation_structured(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	type KV struct {
		Key   string
		Value string
	}

	type Schema struct {
		ID      uint64
		Map     []KV
		Comment string
	}

	schemaUID := registerSchema(t, client, eas.MustNewSchema(Schema{}))

	attestationValues := Schema{
		ID: 3,
		Map: []KV{
			{"k1", "v1"},
			{"k2", "v2"},
		},
		Comment: "Hey",
	}

	_, wait, err := client.EAS.Attest(ctx, schemaUID, nil, attestationValues)
	assertNilError(t, err)

	client.backend.Commit()

	r, err := wait(ctx)
	assertNilError(t, err)

	a, err := client.EAS.GetAttestation(ctx, r.UID)
	assertNilError(t, err)

	assertEqual(t, "schema uid", a.Schema, schemaUID)
	assertEqual(t, "attester", a.Attester, client.account)

	var validationValues Schema
	err = a.ScanValues(&validationValues)

	assertNilError(t, err)
	assertEqual(t, "data", validationValues, attestationValues)
}

func TestEASContract_MultiAttest(t *testing.T) {
	client := newClient(t)
	ctx := context.Background()

	schemaUID := registerSchema(t, client, "string message")

	schemas := [][]any{
		{"one"},
		{"two"},
		{"three"},
		{"four"},
		{"five"},
	}

	_, wait, err := client.EAS.MultiAttest(ctx, schemaUID, &eas.AttestOptions{Revocable: true}, schemas...)
	assertNilError(t, err)

	client.backend.Commit()

	r, err := wait(ctx)
	assertNilError(t, err)

	count := 0
	for i, e := range r {
		a, err := client.EAS.GetAttestation(ctx, e.UID)
		assertNilError(t, err)

		assertNilError(t, err)
		assertEqual(t, "schema uid", e.Schema, schemaUID)
		assertEqual(t, "attester", e.Attester, client.account)

		var message string
		err = a.ScanValues(&message)
		assertNilError(t, err)
		assertEqual(t, "message", message, schemas[i][0].(string))

		count++
	}

	assertEqual(t, "count", count, len(schemas))
}

func attest(t testing.TB, client *Client, schemaUID eas.UID, o *eas.AttestOptions, values ...any) eas.UID {
	t.Helper()

	ctx := context.Background()

	_, wait, err := client.EAS.Attest(ctx, schemaUID, o, values...)
	assertNilError(t, err)

	client.backend.Commit()

	r, err := wait(ctx)
	assertNilError(t, err)

	return r.UID
}
