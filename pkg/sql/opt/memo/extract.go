// Copyright 2018 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package memo

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/types"
)

// This file contains various helper functions that extract an op-specific
// field from an expression.

// ExtractConstDatum returns the Datum that represents the value of an operator
// with a constant value. An operator with a constant value is:
//  - one that has a ConstValue tag, or
//  - a tuple or array where all children are constant values.
func ExtractConstDatum(ev ExprView) tree.Datum {
	switch ev.Operator() {
	case opt.NullOp:
		return tree.DNull

	case opt.TrueOp:
		return tree.DBoolTrue

	case opt.FalseOp:
		return tree.DBoolFalse

	case opt.TupleOp:
		datums := make(tree.Datums, ev.ChildCount())
		for i := range datums {
			datums[i] = ExtractConstDatum(ev.Child(i))
		}
		return tree.NewDTuple(datums...)

	case opt.ArrayOp:
		elementType := ev.Logical().Scalar.Type.(types.TArray).Typ
		a := tree.NewDArray(elementType)
		a.Array = make(tree.Datums, ev.ChildCount())
		for i := range a.Array {
			a.Array[i] = ExtractConstDatum(ev.Child(i))
			if a.Array[i] == tree.DNull {
				a.HasNulls = true
			}
		}
		return a
	}
	if !ev.IsConstValue() {
		panic(fmt.Sprintf("non-const expression: %+v", ev))
	}
	return ev.Private().(tree.Datum)
}
