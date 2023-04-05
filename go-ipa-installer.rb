class GoIpaInstaller < Formula
  desc "Go IPA Installer service for installing iOS apps over HTTP"
  homepage "https://github.com/pandaleecn/go_ipa_installer" # Replace 'pandaleecn' with your actual GitHub username
  url "https://github.com/pandaleecn/go_ipa_installer/archive/refs/tags/v1.0.0.tar.gz" # Replace with the correct URL to the tarball of the release
  sha256 "ff18cccb465a4164e640e02df596a0b624fd0f99d0d4e4cd42422bbad6899af6"
  license "MIT"

  depends_on "go" => :build
  depends_on "libimobiledevice"
  depends_on "ideviceinstaller"

  def install
    bin.install "go_ipa_installer"
  end

  plist_options startup: true, manual: "go_ipa_installer"

  def plist
    <<~EOS
      <?xml version="1.0" encoding="UTF-8"?>
      <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
      <plist version="1.0">
      <dict>
        <key>Label</key>
        <string>#{plist_name}</string>
        <key>ProgramArguments</key>
        <array>
          <string>#{opt_bin}/go_ipa_installer</string>
        </array>
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <true/>
      </dict>
      </plist>
    EOS
  end
end
