// To compile run `aptos move compile`
//
// To get the bytecode `xxd -c 1000000 -p build/multiagent_script/bytecode_scripts/swap.mv`
script {
    use std::signer;
    use aptos_framework::aptos_account;

    /// Swaps Coin1 from party 1 for Coin2 from party 2
    fun swap<Coin1,Coin2>(party1: &signer, party2: &signer, amount1: u64, amount2: u64) {
        let party1_address = signer::address_of(party1);
        let party2_address = signer::address_of(party2);
        aptos_account::transfer_coins<Coin1>(party1, party2_address, amount1);
        aptos_account::transfer_coins<Coin2>(party2, party1_address, amount2);
    }
}
