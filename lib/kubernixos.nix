{ packages ? <nixpkgs>, modules ? [] }:
let
  pkgs = if builtins.isAttrs packages then packages else (import packages {});
  lib = pkgs.lib;

  cfg = (import "${toString pkgs.path}/nixos/lib/eval-config.nix" {
    inherit pkgs modules;
  }).config.kubernixos;

  # Kubernetes label can only have a max length of 63 chars, which explains the substring below
  # see: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
  kubernixos = with builtins; substring 0 62 (hashString "sha256" (toJSON cfg.manifests));

  merge = name: item:
    pkgs.lib.recursiveUpdate item {
      metadata = {
        labels = {
          inherit kubernixos;
        };
      };
    };

  items = lib.mapAttrs merge cfg.manifests;

in
rec{
  eval = {
      config = (removeAttrs cfg ["manifests" "schemas"]) // { checksum = kubernixos; };

      manifests = {
          apiVersion = "v1";
          kind = "List";
          items = lib.attrValues items;
      };
  };

  build =
      pkgs.runCommand "kubernixos-${kubernixos}" { nativeBuiltInputs = [pkgs.kubeval cfg.schemas]; } ''
          mkdir -p $out
          ${lib.concatStringsSep "\n" (lib.mapAttrsToList (n: i: "echo '${builtins.toJSON i}' >$out/${n}.json") items)}
          cd $out
          ${pkgs.kubeval}/bin/kubeval --strict -v ${cfg.version} \
              --output tab \
              --schema-location=file://${cfg.schemas} \
              ${lib.concatMapStringsSep " \\\n" (n: "${n}.json") (lib.attrNames items)}

          ln -s ${pkgs.writeText "kubernixos-${kubernixos}.json" (builtins.toJSON eval.manifests)} $out/all-validated.json
      '';
}
