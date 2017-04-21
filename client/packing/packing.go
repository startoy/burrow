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

// NOTE: this package includes GPLv3 licensed code and CANNOT be published
// under the Apache licensed Hyperledger Burrow code.

package packing

import (
	"fmt"
	"os"

	gethAbi "github.com/ethereum/go-ethereum/accounts/abi"
)

// Inputs is a named map
type Inputs map[string]interface{}

func LoadAbi(abiPath string) (gethAbi.ABI, error) {
	fileAbi, err := os.Open(abiPath)
	if err != nil {
		return gethAbi.ABI{}, fmt.Errorf("Failed to open abi file at %s: %s", abiPath, err)
	}

	loadedAbi, err := gethAbi.JSON(fileAbi)
	if err != nil {
		return gethAbi.ABI{}, fmt.Errorf("Failed to load abi definition at %s: %s", abiPath, err)
	}

	return loadedAbi, nil
}

func GetMethod(abi gethAbi.ABI, name string) (gethAbi.Method, error) {
	method, ok := abi.Methods[name]
	if !ok {
		return gethAbi.Method{}, fmt.Errorf("Failed to find method (name: %s) in abi", name)
	}
	return method, nil
}

// attempt to match the inputs to the method and pack into bytes
// NOTE: pack is only exposed for geth through ABI, so we run in a little
// circle here and pass in ABI to get to ABI.Pack()
func PackInputsForMethod(abi gethAbi.ABI, method gethAbi.Method, inputs Inputs) ([]byte, error) {
	var err error
	args := make([]interface{}, len(method.Inputs))
	for i, argument := range method.Inputs {
		// get value for named argument
		if value, ok := inputs[argument.Name]; ok {
			// JSON decoding has cast value to default types (float, string, bool or nil)
			// attempt to convert value closer to ABI requested input type; errors if
			// conversion not possible
			if value, err = convertToCloserType(&argument.Type, value); err != nil {
				return nil, fmt.Errorf("%s (%s): Error converting type: %s", argument.Name,
					argument.Type.String(), err)
			}
			args[i] = value
		} else {
			return nil, fmt.Errorf("%s (%s): NOT FOUND\n\n", argument.Name, argument.Type.String())
		}
	}
	packedArguments, err := abi.Pack(method.Name, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to pack arguments: %s", err)
	}

	return packedArguments, nil
}
