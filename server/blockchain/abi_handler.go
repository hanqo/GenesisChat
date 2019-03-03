package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"strconv"
	"strings"
)

const abiStr = "[{\"constant\":false,\"inputs\":[{\"name\":\"programURL_\",\"type\":\"string\"}],\"name\":\"setProgramURL\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getName\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getDescription\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getTreasury\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"programAddr_\",\"type\":\"string\"}],\"name\":\"setProgramAddr\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getEntryCost\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"cost_\",\"type\":\"uint256\"}],\"name\":\"setExitCost\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getExitCost\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"getCitizenList\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getProgramAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getProgramURL\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"description_\",\"type\":\"string\"}],\"name\":\"setDescription\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"cost_\",\"type\":\"uint256\"}],\"name\":\"setEntryCost\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"join\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"name_\",\"type\":\"string\"}],\"name\":\"setName\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"leave\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getCitizenList\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"name_\",\"type\":\"string\"},{\"name\":\"description_\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"
const abiBin = "60806040523480156200001157600080fd5b5060405162001bac38038062001bac833981018060405260408110156200003757600080fd5b8101908080516401000000008111156200005057600080fd5b828101905060208101848111156200006757600080fd5b81518560018202830111640100000000821117156200008557600080fd5b50509291906020018051640100000000811115620000a257600080fd5b82810190506020810184811115620000b957600080fd5b8151856001820283011164010000000082111715620000d757600080fd5b505092919050505033600760006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550816000800190805190602001906200013a9291906200015f565b508060006001019080519060200190620001569291906200015f565b5050506200020e565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10620001a257805160ff1916838001178555620001d3565b82800160010185558215620001d3579182015b82811115620001d2578251825591602001919060010190620001b5565b5b509050620001e29190620001e6565b5090565b6200020b91905b8082111562000207576000816000905550600101620001ed565b5090565b90565b61198e806200021e6000396000f3fe6080604052600436106100f6576000357c010000000000000000000000000000000000000000000000000000000090048063049e6125146100fb57806317d7de7c146101c35780631a092541146102535780633b19e84a146102e357806342230c4d1461030e5780635d0f370f146103d65780636769dd9414610401578063695b5d2b1461043c5780636c3a6b4e1461046757806372f67c7c146104e257806382c033fb1461057257806390c3f38f14610602578063ae29f7f9146106ca578063b688a36314610705578063c47f002714610727578063d66d9e19146107ef578063f8b2cb4f14610811578063fa54e99014610876575b600080fd5b34801561010757600080fd5b506101c16004803603602081101561011e57600080fd5b810190808035906020019064010000000081111561013b57600080fd5b82018360208201111561014d57600080fd5b8035906020019184600183028401116401000000008311171561016f57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506108e2565b005b3480156101cf57600080fd5b506101d86108ff565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156102185780820151818401526020810190506101fd565b50505050905090810190601f1680156102455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561025f57600080fd5b506102686109a3565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156102a857808201518184015260208101905061028d565b50505050905090810190601f1680156102d55780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156102ef57600080fd5b506102f8610a48565b6040518082815260200191505060405180910390f35b34801561031a57600080fd5b506103d46004803603602081101561033157600080fd5b810190808035906020019064010000000081111561034e57600080fd5b82018360208201111561036057600080fd5b8035906020019184600183028401116401000000008311171561038257600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610b1a565b005b3480156103e257600080fd5b506103eb610b37565b6040518082815260200191505060405180910390f35b34801561040d57600080fd5b5061043a6004803603602081101561042457600080fd5b8101908080359060200190929190505050610b43565b005b34801561044857600080fd5b50610451610c15565b6040518082815260200191505060405180910390f35b34801561047357600080fd5b506104a06004803603602081101561048a57600080fd5b8101908080359060200190929190505050610c21565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156104ee57600080fd5b506104f7610c64565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561053757808201518184015260208101905061051c565b50505050905090810190601f1680156105645780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561057e57600080fd5b50610587610d09565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156105c75780820151818401526020810190506105ac565b50505050905090810190601f1680156105f45780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561060e57600080fd5b506106c86004803603602081101561062557600080fd5b810190808035906020019064010000000081111561064257600080fd5b82018360208201111561065457600080fd5b8035906020019184600183028401116401000000008311171561067657600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610dae565b005b3480156106d657600080fd5b50610703600480360360208110156106ed57600080fd5b8101908080359060200190929190505050610e90565b005b61070d610f62565b604051808215151515815260200191505060405180910390f35b34801561073357600080fd5b506107ed6004803603602081101561074a57600080fd5b810190808035906020019064010000000081111561076757600080fd5b82018360208201111561077957600080fd5b8035906020019184600183028401116401000000008311171561079b57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050611238565b005b6107f7611319565b604051808215151515815260200191505060405180910390f35b34801561081d57600080fd5b506108606004803603602081101561083457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061169e565b6040518082815260200191505060405180910390f35b34801561088257600080fd5b5061088b6116e7565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b838110156108ce5780820151818401526020810190506108b3565b505050509050019250505060405180910390f35b80600060030190805190602001906108fb929190611891565b5050565b6060600080018054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156109995780601f1061096e57610100808354040283529160200191610999565b820191906000526020600020905b81548152906001019060200180831161097c57829003601f168201915b5050505050905090565b606060006001018054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610a3e5780601f10610a1357610100808354040283529160200191610a3e565b820191906000526020600020905b815481529060010190602001808311610a2157829003601f168201915b5050505050905090565b6000600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515610b0f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636f756e7472792063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b600060040154905090565b8060006002019080519060200190610b33929190611891565b5050565b60008060050154905090565b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515610c08576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636f756e7472792063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b8060006006018190555050565b60008060060154905090565b6000600882815481101515610c3257fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050919050565b606060006002018054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610cff5780601f10610cd457610100808354040283529160200191610cff565b820191906000526020600020905b815481529060010190602001808311610ce257829003601f168201915b5050505050905090565b606060006003018054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610da45780601f10610d7957610100808354040283529160200191610da4565b820191906000526020600020905b815481529060010190602001808311610d8757829003601f168201915b5050505050905090565b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515610e73576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636f756e7472792063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b8060006001019080519060200190610e8c929190611891565b5050565b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515610f55576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636f756e7472792063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b8060006005018190555050565b6000600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415151561102a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636974697a656e2063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b6000600601546000600501540134101515156110fa576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260468152602001807f546865206d6f6e65792073656e74206d757374206265206c617267657220746881526020017f616e207468652073756d206f662074686520656e74727920616e64206578697481526020017f20636f737421000000000000000000000000000000000000000000000000000081525060600191505060405180910390fd5b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc6000600501549081150290604051600060405180830381858888f19350505050158015611167573d6000803e3d6000fd5b506000600501546000600401600082825401925050819055506000600501543403600960003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555060083390806001815401808255809150509060018203906000526020600020016000909192909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550506001905090565b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415156112fd576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636f756e7472792063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b80600080019080519060200190611315929190611891565b5050565b6000600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515156113e1576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f4f6e6c7920636974697a656e2063616e2063616c6c20746869732e000000000081525060200191505060405180910390fd5b34600960003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540192505081905550600060060154600960003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054101515156114ea576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601a8152602001807f4661696c20746f2070617920746865206578697420636f73742100000000000081525060200191505060405180910390fd5b600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc6000600601549081150290604051600060405180830381858888f19350505050158015611557573d6000803e3d6000fd5b50600060060154600960003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254039250508190555060006006015460006004016000828254019250508190555060006115cd33611775565b90506008805490508110156116955760086001600880549050038154811015156115f357fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1660088281548110151561162d57fe5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600880548091906001900361168a9190611911565b50600191505061169b565b60009150505b90565b6000600960008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6060600880548060200260200160405190810160405280929190818152602001828054801561176b57602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311611721575b5050505050905090565b600080600090505b60088054905081101561180d578273ffffffffffffffffffffffffffffffffffffffff166008828154811015156117b057fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161415611800578091505061188c565b808060010191505061177d565b600880549050811415151561188a576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f436974697a656e206973206e6f7420666f756e6421000000000000000000000081525060200191505060405180910390fd5b505b919050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106118d257805160ff1916838001178555611900565b82800160010185558215611900579182015b828111156118ff5782518255916020019190600101906118e4565b5b50905061190d919061193d565b5090565b81548183558181111561193857818360005260206000209182019101611937919061193d565b5b505050565b61195f91905b8082111561195b576000816000905550600101611943565b5090565b9056fea165627a7a72305820f3a9d20230116e6e524e1b42ae2d0ce3b171e39f9c1a9c3f4b9b59bcf4ae22100029"

type ABIHandler struct {
	abiObject abi.ABI
}

func NewABIHandler() *ABIHandler {

	abiObject, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		log.Fatal("Not able to parse ABI string")
		return nil
	}

	h := &ABIHandler{
		abiObject: abiObject,
	}
	return h
}

func (h *ABIHandler) packContractFunc(name string, inputs []string) []byte {

	var res []byte
	var err error

	if err != nil {
		log.Fatal("Not able to parse ABI string")
		return nil
	}

	switch name {
	case "": //deploy contract argv: string(name), string(description), uint(entryCost), uint(exitCost)
		entryCost := big.NewInt(0)
		entryCost, _ = entryCost.SetString(inputs[2], 10)

		exitCost := big.NewInt(0)
		exitCost, _ = exitCost.SetString(inputs[3], 10)

		//TODO: constructor requires 4 inputs instead of 2
		res, err = h.abiObject.Pack("", inputs[0], inputs[1])
		res = append(common.FromHex(abiBin), res...)

	case "join":
		res, err = h.abiObject.Pack(name)

	case "lookup":
		addr := common.HexToAddress(inputs[0])
		res, err = h.abiObject.Pack(name, addr)

	case "leave":
		res, err = h.abiObject.Pack(name)

	case "getCitizenList":
		res, err = h.abiObject.Pack(name)

		//TODO: change in contract from getCitizenList to getCitizen
	case "getCitizen":
		index, err := strconv.ParseUint(inputs[0], 10, 64)
		if err != nil {
			log.Fatal("input string not valid in contract getCitizen")
			return nil
		}
		res, err = h.abiObject.Pack(name, index)

	case "getBalance":
		addr := common.HexToAddress(inputs[0])
		res, err = h.abiObject.Pack(name, addr)

	case "getName":
		res, err = h.abiObject.Pack(name)

	case "setName":
		res, err = h.abiObject.Pack(name, inputs[0])

	case "getDescription":
		res, err = h.abiObject.Pack(name)

	case "setDescription":
		res, err = h.abiObject.Pack(name, inputs[0])

	case "getProgramAddr":
		res, err = h.abiObject.Pack(name)

		//TODO:change func in contract from string to address
	case "setProgramAddr":
		addr := common.HexToAddress(inputs[0])
		res, err = h.abiObject.Pack(name, addr)

	case "getProgramURL":
		res, err = h.abiObject.Pack(name)

	case "setProgramURL":
		res, err = h.abiObject.Pack(name, inputs[0])

	case "getEntryCost":
		res, err = h.abiObject.Pack(name)

	case "setEntryCost":
		entryCost := big.NewInt(0)
		entryCost, _ = entryCost.SetString(inputs[0], 10)
		res, err = h.abiObject.Pack(name, entryCost)

	case "getExitCost":
		res, err = h.abiObject.Pack(name)

	case "setExitCost":
		exitCost := big.NewInt(0)
		exitCost, _ = exitCost.SetString(inputs[0], 10)
		res, err = h.abiObject.Pack(name, exitCost)

	case "getTreasury":

		res, err = h.abiObject.Pack(name)

	default:
		log.Fatal("input function unrecognized")

	}

	if err != nil {
		log.Fatal("Not able to pack contract function")
		return nil
	}
	return res

}

func (h *ABIHandler) unpackContractFunc(data []byte, name string) (string, error) {
	var output string
	var err error

	switch name {
	case "lookup":
		var tmp bool
		err = h.abiObject.Unpack(&tmp, name, data)
		if err != nil {
			log.Fatal("input string not valid in contract getCitizen")
			return "", err
		}

		output = strconv.FormatBool(tmp)

	case "getCitizenList":

		type citizen struct {
			name    string
			balance *big.Int
		}
		var tmp []citizen
		err = h.abiObject.Unpack(&tmp, name, data)
		if err != nil {
			log.Fatal("input string not valid in contract getCitizen")
			return "", err
		}

		btmp, _ := json.Marshal(tmp)
		output = string(btmp)

		//TODO: change in contract from getCitizenList to getCitizen
	case "getCitizen":
		type citizen struct {
			name    string
			balance *big.Int
		}
		var tmp citizen

		err = h.abiObject.Unpack(&tmp, name, data)
		btmp, _ := json.Marshal(tmp)

		output = string(btmp)

	case "getBalance":
		var tmp *big.Int
		err = h.abiObject.Unpack(&tmp, name, data)

		output = tmp.String()

	case "getName":
		err = h.abiObject.Unpack(&output, name, data)

	case "getDescription":
		err = h.abiObject.Unpack(&output, name, data)

	case "getProgramAddr":
		err = h.abiObject.Unpack(&output, name, data)

	case "getProgramURL":
		err = h.abiObject.Unpack(&output, name, data)

	case "getEntryCost":
		var tmp *big.Int
		err = h.abiObject.Unpack(&tmp, name, data)

		output = fmt.Sprint(tmp)

	case "getExitCost":
		var tmp *big.Int
		err = h.abiObject.Unpack(&tmp, name, data)

		output = tmp.String()

	case "getTreasury":

		var tmp *big.Int
		err = h.abiObject.Unpack(&tmp, name, data)

		output = tmp.String()

	default:
		log.Fatal("input function unrecognized")

	}

	if err != nil {
		log.Fatal("Not able to pack contract function")
		return "", err
	}
	return output, nil

}

//
//func packContractFunc(methodName string, params []MsgParam) [] byte{
//
//	abiObj,err:=abi.JSON(strings.NewReader(abiStr))
//	typeRegex := regexp.MustCompile("([a-zA-Z]+)(([0-9]+)(x([0-9]+))?)?")
//
//	if err!=nil {
//		log.Fatal("Not able to parse ABI string")
//		return nil
//	}
//	method,exists := abiObj.Methods[methodName]
//
//	if !exists {
//		log.Fatal("Method not exist")
//		return nil
//	}
//
//	if len(params)!=len(method.Inputs) {
//		log.Fatal("Invalid number of parameters. Method requires %v, but %v provided",len(method.Inputs),len(params))
//		return nil
//	}
//
//	methodParams:=make([]interface{},0,100)
//	for i :=range method.Inputs {
//
//		matches := typeRegex.FindAllStringSubmatch(params[i].typ, -1)
//		if len(matches) == 0 {
//			log.Fatal("invalid type '%v'", params[i].typ)
//		}
//		parsedType := matches[0]
//
//		if strings.Count(params[i].typ, "[") != 0 {
//			strSets :=  strings.Split(params[i].value, " ")
//
//			//TODO(xrisheng): Support 2D Array
//			for _, str := range strSets{
//				tmp, err := forEachStrToValue(parsedType[0], str)
//				if err!=nil {
//					log.Fatal("Failed to convert string %v to value : %v" ,i ,err)
//					return nil
//				}
//				methodParams = append(methodParams,tmp)
//			}
//		} else{
//			v,err:= forEachStrToValue(parsedType[0], params[i].value)
//
//			if err!=nil {
//				log.Fatal("Failed to convert string %v to value : %v" ,i ,err)
//				return nil
//			}
//			methodParams=append(methodParams,v)
//		}
//	}
//
//	bin,err:=abiObj.Pack(methodName, methodParams...)
//
//	if err!=nil {
//		log.Fatal("Cannot convert json ABI to binary")
//		return nil
//	}
//
//	return bin
//}
//
//
//func forEachStrToValue(typ string, strValue string) (v interface{}, err error){
//	param := strings.TrimSpace(strValue)
//
//	switch typ {
//
//	case "string":
//		strVal:=new(string)
//		v =strVal
//		err=json.Unmarshal([]byte(strValue),v)
//
//	case "int", "uint":
//		val:=big.NewInt(0)
//		_,success:=val.SetString(param,10)
//		if !success {
//			log.Fatal("Invalid numeric (base 10) value: %v",param)
//		}
//		v=val
//
//	case "address":
//		if !((len(param)==(common.AddressLength*2+2)) || (len(param)==common.AddressLength*2)) {
//			log.Fatal("Invalid address length (%v), must be 40 (unprefixed) or 42 (prefixed) chars",len(param))
//		} else {
//			var addr common.Address
//			if len(param)==(common.AddressLength*2+2) {
//				addr=common.HexToAddress(param)
//			} else {
//				var data []byte
//				data,err=hex.DecodeString(param)
//				addr.SetBytes(data)
//			}
//			v=addr
//		}
//
//	case "hash":
//		if !((len(param)==(common.HashLength*2+2)) || (len(param)==common.HashLength*2)) {
//			log.Fatal("Invalid hash length, must be 64 (unprefixed) or 66 (prefixed) chars")
//
//		} else {
//			var hash common.Hash
//			if len(param)==(common.HashLength*2+2) {
//				hash=common.HexToHash(param)
//			} else {
//				var data []byte
//				data,err=hex.DecodeString(param)
//				hash.SetBytes(data)
//			}
//			v=hash
//		}
//
//	case "bytes":
//		if len(param)>2 {
//			if (param[0]=='0') && (param[1]=='x') {
//				param=param[2:] // cut 0x prefix
//			}
//		}
//		decodedBytes,_:=hex.DecodeString(param)
//		v=decodedBytes
//
//	case "bool":
//		val:=new(bool)
//		v=val
//		_=json.Unmarshal([]byte(param),v)
//
//	default:
//		log.Fatal("Not supported parameter type: %s", typ)
//	}
//	return v,err
//}
