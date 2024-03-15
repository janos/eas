// Copyright (c) 2024, Janoš Guljaš <janos@resenje.org>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eas

import (
	"errors"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"resenje.org/taint"
)

type SchemaItem struct {
	Name  string
	Type  string
	Value any
}

func decodeAttestationValues(data []byte, schema string) ([]SchemaItem, error) {
	var args abi.Arguments
	var items []SchemaItem
	for _, declaration := range strings.Split(schema, ",") {
		declaration := strings.TrimSpace(declaration)
		parts := strings.Fields(declaration)

		var item SchemaItem
		switch l := len(parts); l {
		case 0, 1:
			continue
		case 2:
			item.Type, item.Name = parts[0], parts[1]
		default:
			item.Type, item.Name = strings.Join(parts[:l-2], " "), parts[l-1]
		}

		t, err := abi.NewType(item.Type, "", nil)
		if err != nil {
			return nil, err
		}

		args = append(args, abi.Argument{
			Type: t,
		})
		items = append(items, item)
	}

	values, err := args.Unpack(data)
	if err != nil {
		return nil, err
	}

	for i, v := range values {
		item := items[i]
		item.Value = v
		items[i] = item
	}
	return items, nil
}

func encodeAttestationValues(values []any) ([]byte, error) {
	var args abi.Arguments

	for _, v := range values {
		typeString, err := getTypeString(v, "")
		if err != nil {
			return nil, err
		}
		t, err := abi.NewType(typeString, "", nil)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{
			Type: t,
		})
	}

	data, err := args.Pack(values...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func scanAttestationValues(data []byte, args ...any) error {
	var abiArgs abi.Arguments

	for _, arg := range args {
		typeString, err := getTypeString(reflect.Indirect(reflect.ValueOf(arg)).Interface(), "")
		if err != nil {
			return err
		}
		t, err := abi.NewType(typeString, "", nil)
		if err != nil {
			return err
		}
		abiArgs = append(abiArgs, abi.Argument{
			Type: t,
		})
	}

	values, err := abiArgs.Unpack(data)
	if err != nil {
		return err
	}

	if len(args) != len(values) {
		return errors.New("unable to unpack all fields")
	}

	for i, arg := range args {
		if err := taint.Inject(values[i], arg); err != nil {
			return err
		}
	}

	return nil
}