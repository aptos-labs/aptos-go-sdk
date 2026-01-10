// Package ans provides Aptos Names Service (ANS) resolution and management.
//
// ANS allows users to use human-readable names (like "alice.apt") instead of
// 64-character hex addresses. This package provides functionality to:
//
//   - Resolve names to addresses
//   - Look up names for addresses (reverse resolution)
//   - Register and manage names
//   - Check name availability
//
// # Basic Usage
//
// Resolve a name to an address:
//
//	client := ans.NewClient(aptosClient)
//	address, err := client.Resolve(ctx, "alice.apt")
//	if err != nil {
//		if errors.Is(err, ans.ErrNameNotFound) {
//			fmt.Println("Name not registered")
//		}
//		return err
//	}
//	fmt.Printf("alice.apt = %s\n", address)
//
// Look up the primary name for an address:
//
//	name, err := client.GetPrimaryName(ctx, address)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("%s is known as %s\n", address, name)
//
// # Name Format
//
// Names follow the format: name.apt or subdomain.name.apt
//
//   - Primary names: "alice.apt", "bob.apt"
//   - Subdomains: "wallet.alice.apt", "games.bob.apt"
//
// # Registration
//
// To register a name, the client provides helper methods:
//
//	// Check availability
//	available, err := client.IsAvailable(ctx, "alice.apt")
//
//	// Register (requires APT payment)
//	result, err := client.Register(ctx, signer, "alice.apt", ans.RegisterOptions{
//		Years: 1,
//	})
//
// # Primary Names
//
// Each address can have a primary name that represents it. This is the
// "canonical" name for that address:
//
//	// Set primary name
//	err := client.SetPrimaryName(ctx, signer, "alice.apt")
//
//	// Get primary name
//	primary, err := client.GetPrimaryName(ctx, address)
package ans
