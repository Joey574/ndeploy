{
  checks.x86_64-linux.fleet-test = pkgs.testers.runNixOSTest {
    name = "ndeploy-test";

    nodes = {

    };

    node1 = { ... }: { imports = [ ./modules/test-node.nix ]; };
    node2 = { ... }: { imports = [ ./modules/test-node.nix ]; };

    testScript = ''
      start_all()
      controller.wait_for_unit("multi-user.target")
      node1.wait_for_unit("sshd.service")
      node2.wait_for_unit("sshd.service")
    '';
  };
}
