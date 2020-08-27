{ config, lib, pkgs, ... }:
let
  cfg = config.kubernixos;

  # Kubernetes label can only have a max length of 63 chars, which explains the substring below
  # see: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
  kubernixos = with builtins; substring 0 62 (hashString "sha256" (toJSON cfg.manifests));

  merge = name: item:
    lib.recursiveUpdate item {
      metadata = {
        labels = {
          inherit kubernixos;
        };
      };
    };

  eval = {
      config = (removeAttrs cfg ["build" "eval" "manifests" "schemas"]) // { checksum = kubernixos; };

      manifests = {
          apiVersion = "v1";
          kind = "List";
          items = lib.mapAttrsToList merge cfg.manifests;
      };
  };

  build =
      pkgs.runCommandLocal "kubernixos-${kubernixos}" { nativeBuiltInputs = [pkgs.kubeval cfg.schemas]; } ''
          mkdir -p $out
          ${lib.concatStringsSep "\n" (lib.mapAttrsToList (n: i: "ln -s ${pkgs.writeText "kubernixos-${n}.json" (builtins.toJSON i)} $out/${n}.json") cfg.manifests)}
          ln -s ${pkgs.writeText "kubernixos-${kubernixos}.json" (builtins.toJSON eval.manifests)} $out/kubernixos.json
          cd $out
          ${pkgs.kubeval}/bin/kubeval --strict -v ${cfg.version} \
              --output tab \
              --schema-location=file://${cfg.schemas} \
              ${lib.concatMapStringsSep " " (kind: "--skip-kinds ${kind}") cfg.skipSchemas} \
              kubernixos.json
      '';
in
{
  options.kubernixos = with lib; with lib.types; {

    manifests = mkOption {
      type = attrsOf attrs;
      default = {};
      description = "Attribute set of kubernetes manifests.";
    };

    schemas = mkOption {
      type = package;
      description = "k8s jsonschemas package";
    };

    version = mkOption {
      type = str;
      description = "Kubernetes version to apply manifest validation against.";
    };

    server = mkOption {
      type = str;
      description = ''
        Address to the kubernetes apiserver.
        This is a safety to ensure that applying kubeconfigs and manifest conf match.
      '';
    };

    eval = mkOption {
      type = attrs;
      description = ''
        Meta attrs that holds the evalutation result of the kubernixos expression
      '';
    };

    build = mkOption {
      type = package;
      description = ''
        The output kubernixos manifest package.
      '';
    };

    skipSchemas = mkOption {
      type = listOf str;
      default = [];
      description = ''
        A list of resource kinds for which manifest validation is skipped.
      '';
      example = [ "CustomResourceDefinition" ];
    };

  };

  config.kubernixos = {
    inherit build eval;
  };
}
