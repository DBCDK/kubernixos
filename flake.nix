{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";

    flake-utils = {
      url = "github:numtide/flake-utils";
    };

    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      pre-commit-hooks,
      nixpkgs,
      flake-utils,
      treefmt-nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        version = "dev";

        pkgs = (import nixpkgs) { inherit system; };

        # Eval the treefmt modules from ./treefmt.nix
        treefmtEval = treefmt-nix.lib.evalModule pkgs ./treefmt.nix;
      in
      {
        # for `nix fmt`
        formatter = treefmtEval.config.build.wrapper;

        # for `nix flake check`
        checks = {
          formatting = treefmtEval.config.build.check self;
          build = self.packages.${system}.default;
          pre-commit-check =
            let
              # some treefmt formatters are not supported in pre-commit-hooks we
              # filter them out for now.
              toFilter = [ "yamlfmt" ];
              filterFn = n: _v: (!builtins.elem n toFilter);
              treefmtFormatters = pkgs.lib.mapAttrs (_n: v: { inherit (v) enable; }) (
                pkgs.lib.filterAttrs filterFn (import ./treefmt.nix).programs
              );
            in
            pre-commit-hooks.lib.${system}.run {
              src = ./.;
              hooks = treefmtFormatters;
            };
        };

        # Accessible through 'nix develop' or 'nix-shell' (legacy)
        devShells.default = import ./shell.nix {
          inherit nixpkgs pkgs;
          kubernixos = self.packages.${system}.default;
          inherit (self.checks.${system}.pre-commit-check) shellHook;
        };

        packages = rec {
          default = kubernixos;
          kubernixos = pkgs.callPackage ./default.nix { inherit pkgs nixpkgs version; };
        };
      }
    )
    // {
      nixosModules = {
        kubernixos = {
          imports = [ ./lib/kubernixos.nix ];
        };
      };
    };
}
