package convert

import (
	"errors"
	"fmt"

	"github.com/dapperlabs/cadence"
	jsoncdc "github.com/dapperlabs/cadence/encoding/json"
	"github.com/dapperlabs/flow-go/crypto"
	"github.com/dapperlabs/flow/protobuf/go/flow/entities"

	"github.com/dapperlabs/flow-go-sdk"
)

var ErrEmptyMessage = errors.New("protobuf message is empty")

func MessageToBlockHeader(m *entities.BlockHeader) flow.BlockHeader {
	return flow.BlockHeader{
		ID:       flow.HashToID(m.GetId()),
		ParentID: flow.HashToID(m.GetParentId()),
		Height:   m.GetHeight(),
	}
}

func BlockHeaderToMessage(b flow.BlockHeader) *entities.BlockHeader {
	return &entities.BlockHeader{
		Id:       b.ID[:],
		ParentId: b.ParentID[:],
		Height:   b.Height,
	}
}

func MessageToTransaction(m *entities.Transaction) (flow.Transaction, error) {
	// TODO: implement new transaction message format
	if m == nil {
		return flow.Transaction{}, ErrEmptyMessage
	}
	return flow.Transaction{}, nil
}

func TransactionToMessage(t flow.Transaction) *entities.Transaction {
	// TODO: implement new transaction message format
	return &entities.Transaction{}
}

func MessageToAccount(m *entities.Account) (flow.Account, error) {
	if m == nil {
		return flow.Account{}, ErrEmptyMessage
	}

	accountKeys := make([]flow.AccountKey, len(m.Keys))
	for i, key := range m.Keys {
		accountKey, err := MessageToAccountKey(key)
		if err != nil {
			return flow.Account{}, err
		}

		accountKeys[i] = accountKey
	}

	return flow.Account{
		Address: flow.BytesToAddress(m.Address),
		Balance: m.Balance,
		Code:    m.Code,
		Keys:    accountKeys,
	}, nil
}

func AccountToMessage(a flow.Account) (*entities.Account, error) {
	accountKeys := make([]*entities.AccountPublicKey, len(a.Keys))
	for i, key := range a.Keys {
		accountKeyMsg, err := AccountKeyToMessage(key)
		if err != nil {
			return nil, err
		}
		accountKeys[i] = accountKeyMsg
	}

	return &entities.Account{
		Address: a.Address.Bytes(),
		Balance: a.Balance,
		Code:    a.Code,
		Keys:    accountKeys,
	}, nil
}

func MessageToAccountKey(m *entities.AccountPublicKey) (flow.AccountKey, error) {
	if m == nil {
		return flow.AccountKey{}, ErrEmptyMessage
	}

	signAlgo := crypto.SigningAlgorithm(m.GetSignAlgo())
	hashAlgo := crypto.HashingAlgorithm(m.GetHashAlgo())

	publicKey, err := crypto.DecodePublicKey(signAlgo, m.GetPublicKey())
	if err != nil {
		return flow.AccountKey{}, err
	}

	return flow.AccountKey{
		PublicKey: publicKey,
		SignAlgo:  signAlgo,
		HashAlgo:  hashAlgo,
		Weight:    int(m.GetWeight()),
	}, nil
}

func AccountKeyToMessage(a flow.AccountKey) (*entities.AccountPublicKey, error) {
	publicKey, err := a.PublicKey.Encode()
	if err != nil {
		return nil, err
	}

	return &entities.AccountPublicKey{
		PublicKey: publicKey,
		SignAlgo:  uint32(a.SignAlgo),
		HashAlgo:  uint32(a.HashAlgo),
		Weight:    uint32(a.Weight),
	}, nil
}

func MessageToEvent(m *entities.Event) (flow.Event, error) {
	value, err := jsoncdc.Decode(m.GetPayload())
	if err != nil {
		return flow.Event{}, nil
	}

	eventValue, isEvent := value.(cadence.Event)
	if !isEvent {
		return flow.Event{}, fmt.Errorf("convert: expected Event value, got %s", eventValue.Type().ID())
	}

	return flow.Event{
		Type:          m.GetType(),
		TransactionID: flow.HashToID(m.GetTransactionId()),
		Index:         uint(m.GetIndex()),
		Value:         eventValue,
	}, nil
}

func EventToMessage(e flow.Event) (*entities.Event, error) {
	payload, err := jsoncdc.Encode(e.Value)
	if err != nil {
		return nil, fmt.Errorf("convert: %w", err)
	}

	return &entities.Event{
		Type:          e.Type,
		TransactionId: e.TransactionID[:],
		Index:         uint32(e.Index),
		Payload:       payload,
	}, nil
}
