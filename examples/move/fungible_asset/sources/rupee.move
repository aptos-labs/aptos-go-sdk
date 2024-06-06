/// Rupee module
///
/// Note this is just for testing, and holds no value, can be burned or transferred away at any time
///
/// To build for publishing run `aptos move build-publish-payload --included-artifacts none --json-output-file fungible_asset.json`
module rupee_addr::rupee {

    use std::option;
    use std::signer;
    use std::string;
    use aptos_framework::fungible_asset;
    use aptos_framework::fungible_asset::{MintRef, TransferRef, BurnRef, Metadata};
    use aptos_framework::object;
    use aptos_framework::object::Object;
    use aptos_framework::primary_fungible_store;

    const RUPEE_SEED: vector<u8> = b"RupeesAreForTesting";
    const RUPEE_NAME: vector<u8> = b"Rupee";
    const RUPEE_SYM: vector<u8> = b"RUPEE";
    const RUPEE_DECIMALS: u8 = 2;
    const RUPEE_URL: vector<u8> = b"ipfs://QmXSBSLo3wDBna31MLVxKJ5oBJHkgoDMxEgjRz9tPDfZPn";

    /// Caller is not owner of the metadata object
    const E_NOT_OWNER: u64 = 1;

    #[resource_group_member(group = aptos_framework::object::ObjectGroup)]
    struct RupeeRefs has key {
        mint_ref: MintRef,
        transfer_ref: TransferRef,
        burn_ref: BurnRef,
        obj_transfer_ref: object::TransferRef,
        obj_extend_ref: object::ExtendRef,
    }

    fun init_module(caller: &signer) {
        initialize(caller)
    }

    fun initialize(caller: &signer) {
        // Create metadata object
        let constructor = object::create_named_object(caller, RUPEE_SEED);

        // Disable transferring the metadata object normally
        let obj_extend_ref = object::generate_extend_ref(&constructor);
        let obj_transfer_ref = object::generate_transfer_ref(&constructor);
        object::disable_ungated_transfer(&obj_transfer_ref);

        // Add fungibility
        primary_fungible_store::create_primary_store_enabled_fungible_asset(
            &constructor,
            option::none(), // No max supply
            string::utf8(RUPEE_NAME),
            string::utf8(RUPEE_SYM),
            RUPEE_DECIMALS,
            string::utf8(RUPEE_URL),
            string::utf8(b"")
        );

        // Store refs
        let mint_ref = fungible_asset::generate_mint_ref(&constructor);
        let transfer_ref = fungible_asset::generate_transfer_ref(&constructor);
        let burn_ref = fungible_asset::generate_burn_ref(&constructor);

        let obj_signer = object::generate_signer(&constructor);
        move_to(&obj_signer, RupeeRefs {
            mint_ref,
            transfer_ref,
            burn_ref,
            obj_transfer_ref,
            obj_extend_ref
        })
    }

    /// Mint assets to an account
    entry fun mint(caller: &signer, receiver: address, amount: u64) acquires RupeeRefs {
        let refs = borrow_metadata(caller);
        primary_fungible_store::mint(&refs.mint_ref, receiver, amount);
    }

    /// Burn assets from an account
    entry fun burn(caller: &signer, receiver: address, amount: u64) acquires RupeeRefs {
        let refs = borrow_metadata(caller);
        primary_fungible_store::burn(&refs.burn_ref, receiver, amount);
    }

    /// Transfer assets from one account to another
    entry fun transfer(caller: &signer, sender: address, receiver: address, amount: u64) acquires RupeeRefs {
        let refs = borrow_metadata(caller);
        primary_fungible_store::transfer_with_ref(&refs.transfer_ref, sender, receiver, amount);
    }

    /// Freeze an account
    entry fun set_freeze(caller: &signer, owner: address, freeze: bool) acquires RupeeRefs {
        let refs = borrow_metadata(caller);
        primary_fungible_store::set_frozen_flag(&refs.transfer_ref, owner, freeze)
    }

    /// Borrow the metadata object
    inline fun borrow_metadata(caller: &signer): &RupeeRefs {
        assert_owner(caller);
        borrow_global<RupeeRefs>(metadata_object_address())
    }

    /// Assert owner of metadata is doing the change
    inline fun assert_owner(caller: &signer) {
        let caller_address = signer::address_of(caller);
        assert!(object::is_owner(metadata_object(), caller_address), E_NOT_OWNER);
    }

    inline fun metadata_object_address(): address {
        object::create_object_address(&@rupee_addr, RUPEE_SEED)
    }

    inline fun metadata_object(): Object<Metadata> {
        object::address_to_object(metadata_object_address())
    }

    #[view]
    public fun fa_address() : address {
        metadata_object_address()
    }
}
