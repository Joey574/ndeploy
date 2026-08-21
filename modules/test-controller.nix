{ pkgs, ndeploy, ... }:

{
  services.xserver.enable = true;
  services.xserver.windowManager.i3.enable = true;

  users.users.controller = {
    isNormalUser = true;
    password = "";
    extraGroups = [ "video" "input" "wheel" ];
  };

  programs.ssh.startAgent = true;
  programs.ssh.askPassword = "${pkgs.x11_ssh_askpass}/libexec/x11-ssh-askpass";

  programs.ssh.extraConfig = ''
    AddKeysToAgent yes
  '';

  systemd.tmpfiles.rules = [
    "d /home/controller/.cache 0755 controller users -"
    "d /home/controller/.config 0755 controller users -"
    "d /home/controller/.ssh 0700 controller users -"

    "C /home/controller/.ssh/id_ed25519 0600 controller users - ${../keys/test_key}"
  ];

  services.xserver.displayManager.lightdm.enable = true;
  services.displayManager.autoLogin = {
    enable = true;
    user = "controller";
  };
  services.displayManager.defaultSession = "none+i3";

  services.xserver.windowManager.i3.configFile = pkgs.writeText "i3-config" ''
    exec ${ndeploy}/bin/ndeploy
    set $mod Mod4
    bindsym $mod+f floating toggle
    bindsym $mod+Shift+q kill
    bindsym $mod+Return exec i3-sensible-terminal
  '';

  virtualisation = {
    graphics = true;
    qemu.options = [
      "-vga none"
      "-device virtio-vga-gl"
      "-display gtk,gl=on,zoom-to-fit=off"
      "-device usb-tablet"
    ];

    cores = 4;
    memorySize = 8192;
    resolution = {
      x = 1920;
      y = 1080;
    };
  };

  environment.systemPackages = with pkgs; [ mesa i3status dmenu openssh ];
  system.stateVersion = "26.05";
}
