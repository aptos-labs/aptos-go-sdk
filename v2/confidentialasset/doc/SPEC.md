# Confidential Asset Go SDK — TS ↔ Move ↔ Go specification

For **TypeScript ↔ Go file and API mapping** (architecture, CGO split, parity gaps), see **[TS_GO_MAP.md](TS_GO_MAP.md)**. This document focuses on **Move** view/entry layouts.

Reference: `@aptos-labs/confidential-asset` (`confidential-asset/src/internal/confidentialAssetTxnBuilder.ts`, `viewFunctions.ts`, `api/confidentialAsset.ts`).  
Module address default: `0x1`, Move module name: `confidential_asset`.

## View functions (TS `viewFunctions` / txn builder)

| Go SDK method (see `views.go`) | Move view | `functionArguments` |
|-------------------------------|-----------|---------------------|
| `GetPendingBalanceCipher` | `{addr}::confidential_asset::get_pending_balance` | `[accountAddress, tokenAddress]` |
| `GetAvailableBalanceCipher` | `{addr}::confidential_asset::get_available_balance` | `[accountAddress, tokenAddress]` |
| `IsBalanceNormalized` | `is_normalized` | `[accountAddress, tokenAddress]` |
| `IncomingTransfersPaused` | `incoming_transfers_paused` | `[accountAddress, tokenAddress]` |
| `HasConfidentialStore` | `has_confidential_store` | `[accountAddress, tokenAddress]` |
| `GetEffectiveAuditorHint` | `get_effective_auditor_hint` | `[accountAddress, tokenAddress]` |
| `IsEmergencyPaused` | `is_emergency_paused` | `[]` |
| `GetEncryptionKey` | `get_encryption_key` | `[accountAddress, tokenAddress]` |
| `GetEffectiveAuditorConfig` | `get_effective_auditor_config` | `[tokenAddress]` |
| `GetMaxMemoBytes` | `get_max_memo_bytes` | `[]` |

Cipher shape: `{ P: [{data: "0x..."}], R: [...] }` per chunk (TS slices `0x` prefix when building `TwistedElGamalCiphertext`).

## Entry functions (TS `ConfidentialAssetTransactionBuilder`)

| TS builder method | Move entry | `functionArguments` (conceptual) |
|-------------------|------------|----------------------------------|
| `registerBalance` | `register_raw` | `[token, ek_bytes, sigma_commitment, sigma_response]` |
| `deposit` | `deposit` | `[token, amount_string]` |
| `withdraw` | `withdraw_to_raw` | `[token, recipient, amount_string, new_C[], new_D[], new_A[], range_proof, sigma_c, sigma_r]` |
| `normalizeBalance` | `normalize_raw` | `[token, new_C[], new_D[], new_A[], range_proof, sigma_c, sigma_r]` |
| `transfer` | `confidential_transfer_raw` | `[token, recipient, new_bal_C[], new_bal_D[], new_bal_A[], sender_amt_C[], sender_amt_D[], recipient_D[], eff_aud_D[], vol_aud_ek[][], vol_aud_D[][], range_new_bal, range_amt, sigma_c, sigma_r, memo]` |
| `rotateEncryptionKey` | `rotate_encryption_key_raw` | `[token, new_ek_bytes, unpause, new_D_bytes, sigma_c, sigma_r]` |
| `rolloverPendingBalance` | `rollover_pending_balance` | `[token]` |
| (pause variant) | `rollover_pending_balance_and_pause` | `[token]` |

## Golden vectors (Phase 0)

Generate with a small TS script (same keys / network as tests): log `len(commitment)`, `len(response)`, `len(rangeProof)`, and hex prefixes for one `register_raw`, `normalize_raw`, `withdraw_to_raw`, `confidential_transfer_raw`. Store under `doc/golden/` when available; Go tests compare layout sizes against TS.

## Chain ID

TS uses `getChainId()` from fullnode for sigma domain separation. Go: use `client.GetLedgerInfo` or equivalent chain id API from aptos-go-sdk v2.
