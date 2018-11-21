{ pkgs }: with pkgs.lib; rec{

  namespace = name: {
    kind = "Namespace";
    apiVersion = "v1";
    metadata = {
      inherit name;
    };
  };

  netpol = { name, namespace, podSelector ? {}, policyTypes ? [ "Ingress" ], ingress ? {}, egress ? {} }: {
    apiVersion = "networking.k8s.io/v1";
    kind = "NetworkPolicy";
    metadata = {
      inherit name namespace;
    };
    spec = {
      inherit podSelector policyTypes;
    } // (optionalAttrs (ingress != {}) { inherit ingress; })
      // (optionalAttrs (egress != {}) { inherit egress; });
  };

}
