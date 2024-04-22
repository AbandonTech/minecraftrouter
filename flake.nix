{
  outputs = { self, nixpkgs, ... }:
    let
      forAllSystems = function:
        nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ]
        (system: function nixpkgs.legacyPackages.${system});
    in rec {
      devShells =
        forAllSystems (pkgs: { default = pkgs.callPackage ./shell.nix { }; });
    };
}
