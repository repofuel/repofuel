// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package metrics

import (
	"fmt"
	"reflect"
)

type ChangeMeasures struct {
	// NS is the number of modified subsystems
	NS float64 `json:"ns"`
	// NS is the number of modified directories
	ND float64 `json:"nd"`
	// NF is the number of modified files
	NF float64 `json:"nf"`
	// Entropy is the distribution of modified code across each file
	Entropy float64 `json:"entropy"`
	// LA is lines of code added in the commit
	LA float64 `json:"la"`
	// LD is lines of code deleted in the commit
	LD float64 `json:"ld"`
	// HA is hunks of code added in the commit
	HA float64 `json:"ha"`
	// HD is hunks of code deleted in the commit
	HD float64 `json:"hd"`
	// LT is lines of code in the modified files before the commit
	LT float64 `json:"lt"`
	// NDEV is the number of developers that changed the modified files in the past
	NDEV float64 `json:"ndev"`
	// AGE is the average time interval between the last and the current change of modified files
	AGE float64 `json:"age"`
	// NUC is the number of unique last changes that touched the modified files
	NUC float64 `json:"nuc"`
	// EXP is the developer experience, measured by the number of submitted commits
	EXP float64 `json:"exp"`
	// REXP is the recent developer experience in the modified files
	REXP float64 `json:"rexp"`
	// SEXP is the developer experience on modified subsystems
	SEXP float64 `json:"sexp"`
}

func (m ChangeMeasures) GetHeaders() []string {
	val := reflect.ValueOf(m)
	n := val.NumField()

	r := make([]string, n)
	for i := 0; i < n; i++ {
		r[i] = val.Type().Field(i).Tag.Get("json")
	}

	return r
}

//todo:
// - The name should indicate that it converted to slice of strings
// - we could just have a another functions
func (m ChangeMeasures) ToSlice() []string {
	val := reflect.ValueOf(m)
	n := val.NumField()

	r := make([]string, n)
	for i := 0; i < n; i++ {
		r[i] = fmt.Sprint(val.Field(i).Interface())
	}

	return r
}

type FileMeasures struct {
	// LA is lines of code added in file
	LA float64 `json:"la"`
	// LD is lines of code deleted in the file
	LD float64 `json:"ld"`
	// HA is hunks of code added in the file
	HA float64 `json:"ha"`
	// HD is hunks of code deleted in the file
	HD float64 `json:"hd"`
	// LT is lines of code in the file before the commit
	LT float64 `json:"lt"`
	// NDEV is the number of developers that changed the file in the past
	NDEV float64 `json:"ndev"`
	// AGE is the time interval between the last and the current change of the file
	AGE float64 `json:"age"`
	// NFC is the number of unique changes that touched the file
	NUC float64 `json:"nuc"`
	// NFC is the number of fix changes
	NFC float64 `json:"nfc"`
	// EXP is developer experience for the file
	EXP float64 `json:"exp"`
	// REXP is the recent developer experience in the file
	REXP float64 `json:"rexp"`
}

type Quantiles struct {
	Commit    map[string]*ChangeMeasures `json:"commit"`
	Developer map[string]*ChangeMeasures `json:"developer"`
	File      map[string]*FileMeasures   `json:"file"`
}

