// To compile run `aptos move compile`
//
// To get the bytecode `xxd -c 1000000 -p build/fa_util_script/bytecode_scripts/transfer.mv`
script {
    use aptos_framework::account;
    use aptos_framework::aptos_account;
    use aptos_framework::object;
    use aptos_framework::primary_fungible_store;
    use aptos_framework::fungible_asset::Metadata;

    fun transfer(caller: &signer, metadata_address: address, receiver: address, amount: u64) {
        let metadata_object = object::address_to_object<Metadata>(metadata_address);
        if (!account::exists_at(receiver)) {
            aptos_account::create_account(receiver)
        };

        primary_fungible_store::transfer(caller, metadata_object, receiver, amount);
    }
}
