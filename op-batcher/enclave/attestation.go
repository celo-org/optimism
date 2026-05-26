package enclave

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

func attest(nonce, userData, publicKey []byte) ([]byte, error) {
	sess, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	res, err := sess.Send(&request.Attestation{
		Nonce:     nonce,
		UserData:  userData,
		PublicKey: publicKey,
	})
	if err != nil {
		return nil, err
	}

	if res.Error != "" {
		return nil, errors.New(string(res.Error))
	}

	if res.Attestation == nil || res.Attestation.Document == nil {
		return nil, errors.New("NSM device did not return an attestation")
	}

	return res.Attestation.Document, nil
}

func AttestationWithPublicKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	// Use empty slices for nonce and publicKey when they're not needed
	nonce := make([]byte, 0)
	txData := make([]byte, 0)
	publicKeyBytes := crypto.FromECDSAPub(publicKey)

	// Call the existing attest function with txData as userData
	return attest(nonce, txData, publicKeyBytes)
}
