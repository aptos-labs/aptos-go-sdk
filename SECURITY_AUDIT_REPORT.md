# Security Audit Report: WebAuthn Signature Support in Aptos Go SDK v2

**Date:** January 30, 2026  
**Auditor:** Security Audit Agent  
**Scope:** WebAuthn signature implementation in `/v2/internal/crypto/webauthn.go`, integration in `/v2/internal/crypto/single_key.go`, and Secp256r1 implementation in `/v2/internal/crypto/secp256r1.go`

---

## Executive Summary

This security audit examined the recently added WebAuthn signature support in the Aptos Go SDK v2. The implementation follows WebAuthn standards and integrates Secp256r1 (P-256) signatures for passkey authentication. While the core cryptographic operations are sound, several security concerns were identified, particularly around timing attack vulnerabilities, input validation, and potential information leakage.

**Overall Risk Assessment:** Medium-High

**Critical Issues:** 1  
**High Issues:** 2  
**Medium Issues:** 3  
**Low Issues:** 2  
**Informational:** 2

---

## 1. Critical Issues

### CRITICAL-1: Timing Attack Vulnerability in Challenge Comparison

**Severity:** Critical  
**CWE:** CWE-208 (Observable Timing Discrepancy)  
**Location:** `v2/internal/crypto/webauthn.go:201-209`

**Description:**
The challenge comparison in `Verify()` and `VerifyArbitraryMessage()` uses a byte-by-byte comparison loop that leaks timing information. An attacker can determine the correct challenge value by measuring response times.

**Vulnerable Code:**
```go
// Verify challenges match
if len(actualChallenge) != len(expectedChallenge) {
    return false
}
for i := range actualChallenge {
    if actualChallenge[i] != expectedChallenge[i] {
        return false
    }
}
```

**Impact:**
- Attackers can perform offline brute-force attacks on challenges
- Enables replay attack construction by learning valid challenge values
- Violates constant-time comparison requirements for cryptographic operations

**Recommendation:**
Use `crypto/subtle.ConstantTimeCompare()` for all challenge comparisons:

```go
import "crypto/subtle"

// Verify challenges match
if len(actualChallenge) != len(expectedChallenge) {
    return false
}
if subtle.ConstantTimeCompare(actualChallenge, expectedChallenge) != 1 {
    return false
}
```

**References:**
- OWASP: Timing Attacks
- CWE-208: Observable Timing Discrepancy
- RFC 8446 Section 7.4.1.3 (constant-time comparisons)

---

## 2. High Severity Issues

### HIGH-1: Missing Bounds Checking on Variable-Length Byte Arrays

**Severity:** High  
**CWE:** CWE-119 (Improper Restriction of Operations within the Bounds of a Memory Buffer)  
**Location:** `v2/internal/crypto/webauthn.go:135-147`

**Description:**
The `UnmarshalBCS()` method for `PartialAuthenticatorAssertionResponse` reads variable-length byte arrays (`AuthenticatorData` and `ClientDataJSON`) without enforcing maximum size limits. While `MaxWebAuthnSignatureBytes` is defined (1024 bytes), it's not enforced during deserialization.

**Vulnerable Code:**
```go
func (p *PartialAuthenticatorAssertionResponse) UnmarshalBCS(des *bcs.Deserializer) {
    des.Struct(&p.Signature)
    if des.Error() != nil {
        return
    }

    p.AuthenticatorData = des.ReadBytes()  // No size limit check
    if des.Error() != nil {
        return
    }

    p.ClientDataJSON = des.ReadBytes()  // No size limit check
}
```

**Impact:**
- Denial of Service (DoS) via memory exhaustion
- Potential integer overflow in length calculations
- Resource exhaustion attacks

**Recommendation:**
Add explicit size checks before deserializing:

```go
func (p *PartialAuthenticatorAssertionResponse) UnmarshalBCS(des *bcs.Deserializer) {
    des.Struct(&p.Signature)
    if des.Error() != nil {
        return
    }

    // Check remaining bytes before reading
    remainingBefore := des.Remaining()
    
    p.AuthenticatorData = des.ReadBytes()
    if des.Error() != nil {
        return
    }
    
    // Enforce maximum size limits
    if len(p.AuthenticatorData) > MaxWebAuthnSignatureBytes {
        des.SetError(fmt.Errorf("authenticator data exceeds maximum size: %d > %d", 
            len(p.AuthenticatorData), MaxWebAuthnSignatureBytes))
        return
    }

    p.ClientDataJSON = des.ReadBytes()
    if des.Error() != nil {
        return
    }
    
    if len(p.ClientDataJSON) > MaxWebAuthnSignatureBytes {
        des.SetError(fmt.Errorf("client data JSON exceeds maximum size: %d > %d", 
            len(p.ClientDataJSON), MaxWebAuthnSignatureBytes))
        return
    }
    
    // Check total size
    totalSize := len(p.AuthenticatorData) + len(p.ClientDataJSON)
    if totalSize > MaxWebAuthnSignatureBytes {
        des.SetError(fmt.Errorf("total WebAuthn data exceeds maximum size: %d > %d", 
            totalSize, MaxWebAuthnSignatureBytes))
        return
    }
}
```

**Note:** The BCS deserializer does perform bounds checking for buffer overflows, but doesn't enforce application-level size limits.

---

### HIGH-2: Insufficient Input Validation in Challenge Decoding

**Severity:** High  
**CWE:** CWE-20 (Improper Input Validation)  
**Location:** `v2/internal/crypto/webauthn.go:149-167`

**Description:**
The `GetChallenge()` method attempts to decode base64url-encoded challenges, falling back to standard base64 if base64url fails. This fallback behavior can mask encoding errors and potentially accept malformed challenges.

**Vulnerable Code:**
```go
func (p *PartialAuthenticatorAssertionResponse) GetChallenge() ([]byte, error) {
    var clientData CollectedClientData
    if err := json.Unmarshal(p.ClientDataJSON, &clientData); err != nil {
        return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
    }

    // Challenge is base64url encoded
    challenge, err := base64.RawURLEncoding.DecodeString(clientData.Challenge)
    if err != nil {
        // Try standard base64
        challenge, err = base64.StdEncoding.DecodeString(clientData.Challenge)
        if err != nil {
            return nil, fmt.Errorf("failed to decode challenge: %w", err)
        }
    }

    return challenge, nil
}
```

**Issues:**
1. Fallback to standard base64 may accept invalid WebAuthn challenges (WebAuthn spec requires base64url)
2. No validation of challenge length (should be 32 bytes for SHA3-256 hash)
3. No validation of challenge format/content

**Impact:**
- Accepts non-standard challenge encodings
- Potential for challenge manipulation attacks
- Non-compliance with WebAuthn specification

**Recommendation:**
Enforce strict base64url decoding and validate challenge format:

```go
func (p *PartialAuthenticatorAssertionResponse) GetChallenge() ([]byte, error) {
    var clientData CollectedClientData
    if err := json.Unmarshal(p.ClientDataJSON, &clientData); err != nil {
        return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
    }

    // WebAuthn spec requires base64url encoding (RFC 4648 Section 5)
    challenge, err := base64.RawURLEncoding.DecodeString(clientData.Challenge)
    if err != nil {
        return nil, fmt.Errorf("failed to decode base64url challenge: %w", err)
    }

    // Validate challenge length (SHA3-256 produces 32-byte hashes)
    if len(challenge) != 32 {
        return nil, fmt.Errorf("invalid challenge length: expected 32 bytes, got %d", len(challenge))
    }

    return challenge, nil
}
```

---

## 3. Medium Severity Issues

### MEDIUM-1: Error Information Leakage

**Severity:** Medium  
**CWE:** CWE-209 (Information Exposure Through an Error Message)  
**Location:** Multiple locations in `webauthn.go`

**Description:**
Error messages reveal internal implementation details that could aid attackers in crafting exploits. The `Verify()` method returns `false` on errors, but error details are lost, while deserialization errors expose structure details.

**Examples:**
```go
// Line 50: Reveals variant value
des.SetError(fmt.Errorf("unknown assertion signature variant: %d", s.Variant))

// Line 152: Reveals JSON parsing details
return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
```

**Impact:**
- Information disclosure about internal state
- Potential for error-based side-channel attacks
- Aids in fuzzing and exploit development

**Recommendation:**
Sanitize error messages in production code:

```go
// For deserialization errors, use generic messages
des.SetError(fmt.Errorf("invalid assertion signature variant"))

// For verification errors, return false without exposing details
// (Already handled correctly in Verify methods)
```

**Note:** The `Verify()` methods correctly return `false` without exposing error details, which is good practice.

---

### MEDIUM-2: Missing Validation of AuthenticatorData Structure

**Severity:** Medium  
**CWE:** CWE-345 (Insufficient Verification of Data Authenticity)  
**Location:** `v2/internal/crypto/webauthn.go:78, 141`

**Description:**
The `AuthenticatorData` field is accepted as raw bytes without validating its structure according to WebAuthn specification. AuthenticatorData should be at least 37 bytes and contain specific fields (RP ID hash, flags, counter).

**Impact:**
- Potential for malformed authenticator data attacks
- Missing validation of authenticator flags (user presence, user verification)
- Counter replay protection not validated

**Recommendation:**
Add validation of AuthenticatorData structure:

```go
// Validate authenticator data structure
func (p *PartialAuthenticatorAssertionResponse) validateAuthenticatorData() error {
    if len(p.AuthenticatorData) < 37 {
        return fmt.Errorf("authenticator data too short: %d < 37", len(p.AuthenticatorData))
    }
    
    // First 32 bytes: RP ID hash
    // Byte 32: Flags
    // Bytes 33-36: Sign count (optional, only if UP flag is set)
    
    flags := p.AuthenticatorData[32]
    
    // Check User Presence (UP) flag
    if flags&0x01 == 0 {
        return fmt.Errorf("user presence flag not set")
    }
    
    // Optionally check User Verification (UV) flag for stronger security
    // if flags&0x04 == 0 {
    //     return fmt.Errorf("user verification flag not set")
    // }
    
    return nil
}

// Call in Verify() before signature verification
func (p *PartialAuthenticatorAssertionResponse) Verify(msg []byte, pubKey *AnyPublicKey) bool {
    // ... existing code ...
    
    // Validate authenticator data structure
    if err := p.validateAuthenticatorData(); err != nil {
        return false
    }
    
    // ... rest of verification ...
}
```

---

### MEDIUM-3: Potential Integer Overflow in ULEB128 Decoding

**Severity:** Medium  
**CWE:** CWE-190 (Integer Overflow or Wraparound)  
**Location:** `v2/internal/bcs/deserializer.go:198-234` (used by WebAuthn deserialization)

**Description:**
While the ULEB128 decoder checks for overflow at the uint32 boundary, the intermediate `result` variable is `uint64`, and the check `result > 0xFFFFFFFF` occurs after potential multiplication operations. However, the implementation appears safe due to the shift limit check.

**Analysis:**
The current implementation has safeguards:
- Shift limit of 35 bits prevents excessive shifts
- Final check ensures result fits in uint32
- Bounds checking in `ReadFixedBytes()` prevents buffer overflows

**Recommendation:**
The current implementation is acceptable, but consider adding explicit checks earlier:

```go
// Current implementation is safe, but could add:
if result > math.MaxUint32 {
    des.SetError(ErrOverflow)
    return 0
}
```

**Status:** Informational - Current implementation appears safe, but worth documenting.

---

## 4. Low Severity Issues

### LOW-1: Missing Origin Validation

**Severity:** Low  
**CWE:** CWE-346 (Origin Validation Error)  
**Location:** `v2/internal/crypto/webauthn.go:56-61`

**Description:**
The `CollectedClientData` structure includes an `Origin` field, but it's never validated during signature verification. While this may be application-specific, WebAuthn best practices recommend origin validation.

**Impact:**
- Potential for cross-origin attacks if not validated at application layer
- Reduced security posture

**Recommendation:**
Document that origin validation should be performed at the application layer, or add optional validation:

```go
// Add to Verify() or create separate method
func (p *PartialAuthenticatorAssertionResponse) ValidateOrigin(expectedOrigin string) error {
    var clientData CollectedClientData
    if err := json.Unmarshal(p.ClientDataJSON, &clientData); err != nil {
        return err
    }
    
    if clientData.Origin != expectedOrigin {
        return fmt.Errorf("origin mismatch: expected %s, got %s", expectedOrigin, clientData.Origin)
    }
    
    return nil
}
```

**Note:** This may be intentionally left to the application layer, which is acceptable if documented.

---

### LOW-2: Missing Type Field Validation

**Severity:** Low  
**CWE:** CWE-20 (Improper Input Validation)  
**Location:** `v2/internal/crypto/webauthn.go:57`

**Description:**
The `Type` field in `CollectedClientData` should be validated to ensure it equals `"webauthn.get"` for assertion responses.

**Recommendation:**
Add validation:

```go
func (p *PartialAuthenticatorAssertionResponse) GetChallenge() ([]byte, error) {
    var clientData CollectedClientData
    if err := json.Unmarshal(p.ClientDataJSON, &clientData); err != nil {
        return nil, fmt.Errorf("failed to parse client data JSON: %w", err)
    }
    
    // Validate type field
    if clientData.Type != "webauthn.get" {
        return nil, fmt.Errorf("invalid client data type: expected 'webauthn.get', got '%s'", clientData.Type)
    }
    
    // ... rest of function ...
}
```

---

## 5. Informational Issues

### INFO-1: MaxWebAuthnSignatureBytes Constant Not Enforced

**Severity:** Informational  
**Location:** `v2/internal/crypto/webauthn.go:16`

**Description:**
The constant `MaxWebAuthnSignatureBytes = 1024` is defined but not used anywhere in the codebase. This should be enforced during deserialization (see HIGH-1).

**Recommendation:**
Enforce this limit as described in HIGH-1.

---

### INFO-2: Documentation of Challenge Hash Algorithm

**Severity:** Informational  
**Location:** `v2/internal/crypto/webauthn.go:66, 187`

**Description:**
The documentation states that the challenge should be `SHA3-256(signing_message(transaction))`, which is correctly implemented. However, the comment could be more explicit about the exact format.

**Recommendation:**
Enhance documentation:

```go
// PartialAuthenticatorAssertionResponse contains a subset of the fields
// from a WebAuthn AuthenticatorAssertionResponse.
//
// The challenge in client_data_json must be base64url-encoded SHA3-256 hash
// of the signing message: base64url(SHA3-256(signing_message(transaction)))
//
// The signing_message is typically the BCS-serialized transaction with
// domain separator prefix.
```

---

## 6. Cryptographic Correctness Analysis

### Signature Verification Flow

The signature verification process follows WebAuthn standards:

1. ✅ **Challenge Extraction:** Correctly extracts and decodes base64url challenge
2. ✅ **Challenge Verification:** Compares against SHA3-256 hash of message (needs constant-time comparison - CRITICAL-1)
3. ✅ **Verification Data Construction:** Correctly constructs `authenticator_data || SHA-256(client_data_json)`
4. ✅ **Signature Verification:** Uses ECDSA verification with SHA-256 over verification data

### Secp256r1 Implementation

The Secp256r1 implementation (`secp256r1.go`) is cryptographically sound:

1. ✅ **Key Generation:** Uses `crypto/rand` and `ecdsa.GenerateKey()` with P-256 curve
2. ✅ **Signature Normalization:** Correctly normalizes `s` to low order (prevents signature malleability)
3. ✅ **Signature Validation:** Validates `r` and `s` are in valid range [1, n-1]
4. ✅ **Low S Enforcement:** Enforces low `s` values during deserialization
5. ✅ **Public Key Validation:** Validates public key is on curve and in uncompressed format

### Integration Points

The integration in `single_key.go` correctly routes WebAuthn signatures:

1. ✅ **Variant Detection:** Properly identifies `AnySignatureVariantWebAuthn`
2. ✅ **Type Assertion:** Correctly casts to `PartialAuthenticatorAssertionResponse`
3. ✅ **Verification Routing:** Routes to WebAuthn-specific verification logic

---

## 7. BCS Serialization/Deserialization Safety

### Deserializer Safety

The BCS deserializer (`v2/internal/bcs/deserializer.go`) implements proper bounds checking:

1. ✅ **Buffer Bounds:** All read operations check `des.pos + length <= len(des.source)`
2. ✅ **Error Propagation:** Errors are properly set and checked before subsequent operations
3. ✅ **ULEB128 Limits:** Prevents excessive shifts and overflow
4. ⚠️ **Size Limits:** Application-level size limits not enforced (see HIGH-1)

### WebAuthn Deserialization

The WebAuthn deserialization properly handles errors:

1. ✅ **Early Returns:** Checks errors after each deserialization step
2. ✅ **Variant Validation:** Validates signature variant before deserializing
3. ⚠️ **Size Limits:** Missing maximum size enforcement (HIGH-1)

---

## 8. Memory Safety

### Sensitive Data Handling

1. ✅ **Private Keys:** Secp256r1 private keys are properly redacted in `String()` method
2. ✅ **No Logging:** No obvious logging of sensitive data
3. ⚠️ **Challenge Exposure:** Challenges are stored in memory but not explicitly cleared (acceptable for Go's GC)

### Buffer Management

1. ✅ **Copy Operations:** Uses `copy()` for safe buffer operations
2. ✅ **Bounds Checking:** BCS deserializer prevents buffer overflows
3. ✅ **Slice Safety:** No unsafe pointer arithmetic

---

## 9. Replay Attack Prevention

### Current Implementation

The current implementation does not include explicit replay attack prevention mechanisms:

1. ⚠️ **No Nonce Tracking:** Challenges are not tracked to prevent reuse
2. ⚠️ **No Timestamp Validation:** No validation of challenge freshness
3. ✅ **Challenge Binding:** Challenges are cryptographically bound to transactions

### Recommendations

Replay attack prevention should be implemented at the application layer:

1. **Challenge Tracking:** Maintain a set of used challenges with expiration
2. **Timestamp Validation:** Include timestamp in challenge or validate freshness
3. **One-Time Use:** Ensure each challenge can only be used once

**Note:** This is acceptable if handled at the application layer, but should be documented.

---

## 10. Recommendations Summary

### Immediate Actions (Critical/High)

1. **CRITICAL:** Replace byte-by-byte challenge comparison with `crypto/subtle.ConstantTimeCompare()`
2. **HIGH:** Add maximum size limits to `AuthenticatorData` and `ClientDataJSON` deserialization
3. **HIGH:** Enforce strict base64url decoding and validate challenge length (32 bytes)

### Short-Term Actions (Medium)

4. **MEDIUM:** Add AuthenticatorData structure validation
5. **MEDIUM:** Sanitize error messages to prevent information leakage
6. **MEDIUM:** Add validation for `Type` field in `CollectedClientData`

### Long-Term Actions (Low/Informational)

7. **LOW:** Document origin validation requirements
8. **INFO:** Enforce `MaxWebAuthnSignatureBytes` constant
9. **INFO:** Enhance documentation with explicit format specifications

---

## 11. Testing Recommendations

### Security Test Cases

1. **Timing Attack Tests:**
   - Measure verification time for correct vs incorrect challenges
   - Verify constant-time comparison implementation

2. **Input Validation Tests:**
   - Test with oversized `AuthenticatorData` and `ClientDataJSON`
   - Test with malformed base64/base64url challenges
   - Test with invalid challenge lengths

3. **Deserialization Tests:**
   - Test with truncated BCS data
   - Test with invalid variant values
   - Test with malformed JSON in `ClientDataJSON`

4. **Cryptographic Tests:**
   - Verify signature verification correctness
   - Test with edge cases (all zeros, all ones)
   - Test signature malleability prevention

---

## 12. Compliance and Standards

### WebAuthn Specification Compliance

- ✅ Follows WebAuthn assertion response structure
- ✅ Correctly constructs verification data
- ✅ Uses SHA-256 for client data hash
- ⚠️ Missing AuthenticatorData structure validation
- ⚠️ Missing origin validation (may be application-level)

### Cryptographic Standards

- ✅ Uses NIST P-256 (secp256r1) curve
- ✅ Follows ECDSA signature format
- ✅ Implements signature normalization (low s)
- ✅ Uses secure random number generation

---

## Conclusion

The WebAuthn signature implementation in the Aptos Go SDK v2 is generally well-designed and follows cryptographic best practices. The core cryptographic operations are sound, and the integration with the existing SDK architecture is clean.

However, the **critical timing attack vulnerability** in challenge comparison must be addressed immediately, as it poses a significant security risk. The high-severity issues around input validation and bounds checking should also be prioritized.

With the recommended fixes, this implementation will be production-ready and secure for use in WebAuthn-based authentication flows.

---

## Appendix: Code References

### Files Audited

1. `/v2/internal/crypto/webauthn.go` - Main WebAuthn implementation
2. `/v2/internal/crypto/single_key.go` - Integration layer
3. `/v2/internal/crypto/secp256r1.go` - Secp256r1 cryptographic primitives
4. `/v2/internal/bcs/deserializer.go` - BCS deserialization (security context)
5. `/v2/internal/util/util.go` - Utility functions (SHA3-256)

### Key Functions

- `PartialAuthenticatorAssertionResponse.Verify()` - Main verification logic
- `PartialAuthenticatorAssertionResponse.GetChallenge()` - Challenge extraction
- `PartialAuthenticatorAssertionResponse.generateVerificationData()` - Data construction
- `PartialAuthenticatorAssertionResponse.verifySecp256r1Raw()` - Signature verification
- `Secp256r1Signature.FromBytes()` - Signature validation
- `bcs.Deserializer.ReadBytes()` - Variable-length byte array reading

---

**Report Generated:** January 30, 2026  
**Next Review:** After critical and high-severity issues are addressed
