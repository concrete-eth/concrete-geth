// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package scwallet

import (
	//"crypto/ecdsa"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	//"math/big"
	"github.com/ebfe/scard"
	"github.com/ethereum/go-ethereum/crypto"
	ecdh "github.com/wsddn/go-ecdh"
)

const (
	MAX_PAYLOAD_SIZE   = 223
	PAIR_P1_FIRST_STEP = 0
	PAIR_P1_LAST_STEP  = 1

	SC_SECRET_LENGTH = 32
	SC_BLOCK_SIZE    = 16
)

// SecureChannelSession enables secure communication with a hardware wallet
type SecureChannelSession struct {
	card          *scard.Card // A handle to the smartcard for communication
	secret        []byte      // A shared secret generated from our ECDSA keys
	publicKey     []byte      // Our own ephemeral public key
	PairingKey    []byte      // A permanent shared secret for a pairing, if present
	sessionEncKey []byte      // The current session encryption key
	sessionMacKey []byte      // The current session MAC key
	iv            []byte      // The current IV
	PairingIndex  uint8       // The pairing index
}

// NewSecureChannelSession creates a new secure channel for the given card and public key
func NewSecureChannelSession(card *scard.Card, keyData []byte) (*SecureChannelSession, error) {
	// Generate an ECDSA keypair for ourselves
	gen := ecdh.NewEllipticECDH(crypto.S256())
	private, public, err := gen.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	cardPublic, ok := gen.Unmarshal(keyData)
	if !ok {
		return nil, fmt.Errorf("Could not unmarshal public key from card")
	}

	secret, err := gen.GenerateSharedSecret(private, cardPublic)
	if err != nil {
		return nil, err
	}

	return &SecureChannelSession{
		card:      card,
		secret:    secret,
		publicKey: gen.Marshal(public),
	}, nil
}

// Pair establishes a new pairing with the smartcard
func (s *SecureChannelSession) Pair(sharedSecret []byte) error {
	secretHash := sha256.Sum256(sharedSecret)

	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return err
	}

	response, err := s.pair(PAIR_P1_FIRST_STEP, challenge)
	if err != nil {
		return err
	}

	md := sha256.New()
	md.Write(secretHash[:])
	md.Write(challenge)

	expectedCryptogram := md.Sum(nil)
	cardCryptogram := response.Data[:32]
	cardChallenge := response.Data[32:]

	if !bytes.Equal(expectedCryptogram, cardCryptogram) {
		return fmt.Errorf("Invalid card cryptogram")
	}

	md.Reset()
	md.Write(secretHash[:])
	md.Write(cardChallenge)
	response, err = s.pair(PAIR_P1_LAST_STEP, md.Sum(nil))
	if err != nil {
		return err
	}

	md.Reset()
	md.Write(secretHash[:])
	md.Write(response.Data[1:])
	s.PairingKey = md.Sum(nil)
	s.PairingIndex = response.Data[0]

	return nil
}

// Unpair disestablishes an existing pairing
func (s *SecureChannelSession) Unpair() error {
	if s.PairingKey == nil {
		return fmt.Errorf("Cannot unpair: not paired")
	}

	_, err := s.TransmitEncrypted(CLA_SCWALLET, INS_UNPAIR, s.PairingIndex, 0, []byte{})
	if err != nil {
		return err
	}
	s.PairingKey = nil
	// Close channel
	s.iv = nil
	return nil
}

// Open initializes the secure channel
func (s *SecureChannelSession) Open() error {
	if s.iv != nil {
		return fmt.Errorf("Session already opened")
	}

	response, err := s.open()
	if err != nil {
		return err
	}

	// Generate the encryption/mac key by hashing our shared secret,
	// pairing key, and the first bytes returned from the Open APDU.
	md := sha512.New()
	md.Write(s.secret)
	md.Write(s.PairingKey)
	md.Write(response.Data[:SC_SECRET_LENGTH])
	keyData := md.Sum(nil)
	s.sessionEncKey = keyData[:SC_SECRET_LENGTH]
	s.sessionMacKey = keyData[SC_SECRET_LENGTH : SC_SECRET_LENGTH*2]

	// The IV is the last bytes returned from the Open APDU.
	s.iv = response.Data[SC_SECRET_LENGTH:]

	if err := s.mutuallyAuthenticate(); err != nil {
		return err
	}

	return nil
}

// mutuallyAuthenticate is an internal method to authenticate both ends of the
// connection.
func (s *SecureChannelSession) mutuallyAuthenticate() error {
	data := make([]byte, SC_SECRET_LENGTH)
	if _, err := rand.Read(data); err != nil {
		return err
	}

	response, err := s.TransmitEncrypted(CLA_SCWALLET, INS_MUTUALLY_AUTHENTICATE, 0, 0, data)
	if err != nil {
		return err
	}
	if response.Sw1 != 0x90 || response.Sw2 != 0x00 {
		return fmt.Errorf("Got unexpected response from MUTUALLY_AUTHENTICATE: 0x%x%x", response.Sw1, response.Sw2)
	}

	if len(response.Data) != SC_SECRET_LENGTH {
		return fmt.Errorf("Response from MUTUALLY_AUTHENTICATE was %d bytes, expected %d", len(response.Data), SC_SECRET_LENGTH)
	}

	return nil
}

// open is an internal method that sends an open APDU
func (s *SecureChannelSession) open() (*ResponseAPDU, error) {
	return transmit(s.card, &CommandAPDU{
		Cla:  CLA_SCWALLET,
		Ins:  INS_OPEN_SECURE_CHANNEL,
		P1:   s.PairingIndex,
		P2:   0,
		Data: s.publicKey,
		Le:   0,
	})
}

// pair is an internal method that sends a pair APDU
func (s *SecureChannelSession) pair(p1 uint8, data []byte) (*ResponseAPDU, error) {
	return transmit(s.card, &CommandAPDU{
		Cla:  CLA_SCWALLET,
		Ins:  INS_PAIR,
		P1:   p1,
		P2:   0,
		Data: data,
		Le:   0,
	})
}

// TransmitEncrypted sends an encrypted message, and decrypts and returns the response
func (s *SecureChannelSession) TransmitEncrypted(cla, ins, p1, p2 byte, data []byte) (*ResponseAPDU, error) {
	if s.iv == nil {
		return nil, fmt.Errorf("Channel not open")
	}

	data, err := s.encryptAPDU(data)
	if err != nil {
		return nil, err
	}
	meta := []byte{cla, ins, p1, p2, byte(len(data) + SC_BLOCK_SIZE), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if err = s.updateIV(meta, data); err != nil {
		return nil, err
	}

	fulldata := make([]byte, len(s.iv)+len(data))
	copy(fulldata, s.iv)
	copy(fulldata[len(s.iv):], data)

	response, err := transmit(s.card, &CommandAPDU{
		Cla:  cla,
		Ins:  ins,
		P1:   p1,
		P2:   p2,
		Data: fulldata,
	})
	if err != nil {
		return nil, err
	}

	rmeta := []byte{byte(len(response.Data)), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rmac := response.Data[:len(s.iv)]
	rdata := response.Data[len(s.iv):]
	plainData, err := s.decryptAPDU(rdata)
	if err != nil {
		return nil, err
	}

	if err = s.updateIV(rmeta, rdata); err != nil {
		return nil, err
	}
	if !bytes.Equal(s.iv, rmac) {
		return nil, fmt.Errorf("Invalid MAC in response")
	}

	rapdu := &ResponseAPDU{}
	rapdu.deserialize(plainData)

	if rapdu.Sw1 != SW1_OK {
		return nil, fmt.Errorf("Unexpected response status Cla=0x%x, Ins=0x%x, Sw=0x%x%x", cla, ins, rapdu.Sw1, rapdu.Sw2)
	}

	return rapdu, nil
}

// encryptAPDU is an internal method that serializes and encrypts an APDU
func (s *SecureChannelSession) encryptAPDU(data []byte) ([]byte, error) {
	if len(data) > MAX_PAYLOAD_SIZE {
		return nil, fmt.Errorf("Payload of %d bytes exceeds maximum of %d", len(data), MAX_PAYLOAD_SIZE)
	}
	data = pad(data, 0x80)

	ret := make([]byte, len(data))

	a, err := aes.NewCipher(s.sessionEncKey)
	if err != nil {
		return nil, err
	}
	crypter := cipher.NewCBCEncrypter(a, s.iv)
	crypter.CryptBlocks(ret, data)
	return ret, nil
}

// pad applies message padding to a 16 byte boundary
func pad(data []byte, terminator byte) []byte {
	padded := make([]byte, (len(data)/16+1)*16)
	copy(padded, data)
	padded[len(data)] = terminator
	return padded
}

// decryptAPDU is an internal method that decrypts and deserializes an APDU
func (s *SecureChannelSession) decryptAPDU(data []byte) ([]byte, error) {
	a, err := aes.NewCipher(s.sessionEncKey)
	if err != nil {
		return nil, err
	}

	ret := make([]byte, len(data))

	crypter := cipher.NewCBCDecrypter(a, s.iv)
	crypter.CryptBlocks(ret, data)
	return unpad(ret, 0x80)
}

// unpad strips padding from a message
func unpad(data []byte, terminator byte) ([]byte, error) {
	for i := 1; i <= 16; i++ {
		switch data[len(data)-i] {
		case 0:
			continue
		case terminator:
			return data[:len(data)-i], nil
		default:
			return nil, fmt.Errorf("Expected end of padding, got %d", data[len(data)-i])
		}
	}
	return nil, fmt.Errorf("Expected end of padding, got 0")
}

// updateIV is an internal method that updates the initialization vector after
// each message exchanged.
func (s *SecureChannelSession) updateIV(meta, data []byte) error {
	data = pad(data, 0)
	a, err := aes.NewCipher(s.sessionMacKey)
	if err != nil {
		return err
	}
	crypter := cipher.NewCBCEncrypter(a, make([]byte, 16))
	crypter.CryptBlocks(meta, meta)
	crypter.CryptBlocks(data, data)
	// The first 16 bytes of the last block is the MAC
	s.iv = data[len(data)-32 : len(data)-16]
	return nil
}
