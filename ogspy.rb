class Ogspy < Formula
  desc "CLI to inspect, validate and monitor Open Graph metadata"
  homepage "https://github.com/vincenzomaritato/ogspy"
  version "v1.0.0"
  else
    url "https://github.com/vincenzomaritato/ogspy/releases/download/v1.0.0/ogspy-v1.0.0-darwin-amd64"
    sha256 "a28eb3191847a72f70fd2a06ec7173f5653e757cd3d2af18183fa35344fec8ae"
  if Hardware::CPU.arm?
    url "https://github.com/vincenzomaritato/ogspy/releases/download/v1.0.0/ogspy-v1.0.0-darwin-arm64"
    sha256 "c5007cff371acc6a03490e680285c2994ccff2a7badc5ed6cee781f6535aa7a2"
  end

  def install
    bin.install Dir["ogspy*"][0] => "ogspy"
  end
end
