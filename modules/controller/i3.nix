{ pkgs, ... }:

{
  services.displayManager.defaultSession = "none+i3";

  services.xserver.windowManager.i3.configFile = pkgs.writeText "i3-config" ''
    set $mod Mod4
    bindsym $mod+f floating toggle
    bindsym $mod+Shift+q kill
    bindsym $mod+Return exec i3-sensible-terminal
    bindsym $mod+d exec --no-startup-id dmenu-run
  '';

  environment.systemPackages = with pkgs; [
    i3status dmenu
  ];
}
