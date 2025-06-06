class Ogspy < Formula
  desc      "CLI to inspect, validate and monitor Open Graph metadata"
  homepage  "https://github.com/vincenzomaritato/ogspy"
  url       "https://github.com/vincenzomaritato/ogspy/releases/download/v1.0.0/ogspy_1.0.0_macOS_amd64.tar.gz"
  sha256    "PUT_REAL_SHA256_HERE"
  license   "MIT"

  def install
    bin.install "ogspy"
    generate_completions_from_executable(bin/"ogspy", "completion")
  end

  test do
    assert_match "Lightweight CLI tool", shell_output("#{bin}/ogspy --help")
  end
end