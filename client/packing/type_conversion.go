// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packing

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	gethAbi "github.com/ethereum/go-ethereum/accounts/abi"
)

// castToCloserType tries to recast golang type to golang type that is
// closer to the desired ABI type, errors if casting fails or conversion not defined.
func convertToCloserType(inputType *gethAbi.Type, argument interface{}) (interface{}, error) {
	switch inputType.T {
	case gethAbi.IntTy:
		return convertToInt(argument, inputType.Size)
	case gethAbi.UintTy:
		return convertToUint(argument, inputType.Size)
		// case gethAbi.BoolTy:
	case gethAbi.StringTy:
		return convertToString(argument)
		// case gethAbi.SliceTy:
		// case gethAbi.AddressTy:
		// case gethAbi.FixedBytesTy:
		// case gethAbi.BytesTy:
		// 	return convertToBytes(argument)
		// case gethAbi.HashTy:
		// case gethAbi.FixedpointTy:
		// case gethAbi.FunctionTy:
		// default:
	}
	return nil, nil
}

// convertToInt is idempotent for int; for other types
// it tries to convert the value to var sized int, or fails
func convertToInt(argument interface{}, size int) (interface{}, error) {
	switch t := argument.(type) {
	case int, int8, int16, int32, int64:
		y, ok := t.(int64)
		if !ok {
			return nil, fmt.Errorf("Failed to assert intX as int64")
		}
		// TODO: not tested this works, or makes sense; currently un-used code path
		return reduceToVarSizeInt(y, size)
	case uint: // ignore uintptr for now
		// avoid overrunning
		if t <= math.MaxInt64 {
			return reduceToVarSizeInt(int64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert uint to int: bigger than max int64")
		}
	case uint8:
		return reduceToVarSizeInt(int64(t), size)
	case uint16:
		return reduceToVarSizeInt(int64(t), size)
	case uint32:
		return reduceToVarSizeInt(int64(t), size)
	case uint64:
		if t <= math.MaxInt64 {
			return reduceToVarSizeInt(int64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert uint to int: bigger than max int64")
		}
	case float32:
		var y int64
		// float can be significantly bigger than int
		if math.Abs(float64(t)) <= math.MaxInt64 {
			y = int64(t)
			if float32(y) == t {
				return reduceToVarSizeInt(y, size)
			} else {
				return nil, fmt.Errorf("Failed to convert float32 to int: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float32 to int: bigger than max int64")
		}
	case float64:
		var y int64
		// float can be significantly bigger than int
		if math.Abs(t) <= math.MaxInt64 {
			y = int64(t)
			if float64(y) == t {
				return reduceToVarSizeInt(y, size)
			} else {
				return nil, fmt.Errorf("Failed to convert float64 to int: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float64 to int: bigger than max int64")
		}
	case string:
		y, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil, err
		}
		return reduceToVarSizeInt(y, size)
	case bool:
		if t {
			return reduceToVarSizeInt(int64(1), size)
		} else {
			return reduceToVarSizeInt(int64(0), size)
		}
	case complex64, complex128:
		return nil, fmt.Errorf("Failed to convert complex type to int")
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to int")
	}
}

// this is logic we do not want to keep but it solves the problem
// up to 64bit int for now; definitely in need of better solution
func reduceToVarSizeInt(integer int64, size int) (interface{}, error) {
	// for now map to golang type sizes + big.Int for 256bits
	switch size {
	case 8:
		var x int8
		x = int8(integer)
		if int64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce int64 to int8: overflow")
		}
	case 16:
		var x int16
		x = int16(integer)
		if int64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce int64 to int16: overflow")
		}
	case 32:
		var x int32
		x = int32(integer)
		if int64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce int64 to int32: overflow")
		}
	case 64:
		return integer, nil
	case 256:
		i := new(big.Int)
		i.SetInt64(integer)
		return i, nil
	case -1:
		return nil, fmt.Errorf("Failed to reduce int64: size undefined")
	default:
		return nil, fmt.Errorf("Failed to reduce int64: size %v unhandled", size)
	}
}

// convertToUint is idempotent for uint; for other types
// it tries to convert the value to uint64, or fails
func convertToUint(argument interface{}, size int) (interface{}, error) {
	switch t := argument.(type) {
	case int:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int8:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int16:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int32:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int64:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case uint, uint8, uint16, uint32, uint64: // ignore uintptr for now
		y, ok := t.(uint64)
		if !ok {
			return nil, fmt.Errorf("Failed to assert uintX as uint64")
		}
		return reduceToVarSizeUint(y, size)
	case float32:
		var y uint64
		// float can be significantly bigger than int
		if math.Abs(float64(t)) <= math.MaxUint64 && t >= 0 {
			y = uint64(t)
			if float32(y) == t {
				return reduceToVarSizeUint(uint64(y), size)
			} else {
				return nil, fmt.Errorf("Failed to convert float32 to uint: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float32 to uint: bigger than max uint64 or negative")
		}
	case float64:
		var y uint64
		// float can be significantly bigger than int
		if math.Abs(t) <= math.MaxUint64 && t >= 0 {
			y = uint64(t)
			if float64(y) == t {
				return reduceToVarSizeUint(uint64(y), size)
			} else {
				return nil, fmt.Errorf("Failed to convert float64 to uint: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float64 to uint: bigger than max uint64 or negative")
		}
	case string:
		y, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return nil, err
		}
		return reduceToVarSizeUint(y, size)
	case bool:
		if t {
			return reduceToVarSizeUint(uint64(1), size)
		} else {
			return reduceToVarSizeUint(uint64(0), size)
		}
	case complex64, complex128:
		return nil, fmt.Errorf("Failed to convert complex type to uint")
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to uint")
	}
}

// this is logic we do not want to keep but it solves the problem
// up to 64bit uint for now; definitely in need of better solution
func reduceToVarSizeUint(integer uint64, size int) (interface{}, error) {
	// for now map to golang type sizes + big.Int for 256bits
	switch size {
	case 8:
		var x uint8
		x = uint8(integer)
		if uint64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce uint64 to uint8: overflow")
		}
	case 16:
		var x uint16
		x = uint16(integer)
		if uint64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce uint64 to uint16: overflow")
		}
	case 32:
		var x uint32
		x = uint32(integer)
		if uint64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce uint64 to uint32: overflow")
		}
	case 64:
		return integer, nil
	case 256:
		i := new(big.Int)
		i.SetUint64(integer)
		return i, nil
	case -1:
		return nil, fmt.Errorf("Failed to reduce uint64: size undefined")
	default:
		return nil, fmt.Errorf("Failed to reduce uint64: size %v unhandled", size)
	}
}

// convertToString is idempotent for string; for other types it fails
// can be extended for other type
func convertToString(argument interface{}) (interface{}, error) {
	switch argument.(type) {
	case string:
		return argument, nil
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to string")
	}
}
