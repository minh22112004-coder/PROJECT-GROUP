import re
from pathlib import Path

LOG_DIR = Path("logs")


def count_stages_and_anchors(log_path: Path):
    """Count Stage and Anchor signals from dynamic-analysis logs.

    Stages: number of occurrences of:
      - "Testing:"
      - "BƯỚC"
      - "target:" (case-insensitive, supports target\s*:\s*)

    Anchors: number of occurrences of HTTP status markers:
      - "Status: <code>"

    Note: This parser is intentionally simple and aligns with the existing
    test scripts' printed markers.
    """
    if not log_path.exists():
        return None, None

    text = log_path.read_text(encoding="utf-8", errors="ignore")

    # Stages markers
    stages = len(re.findall(r"Testing:|BƯỚC|target\s*:", text, re.IGNORECASE))

    # Anchors marker
    anchors = len(re.findall(r"Status:\s*\d{3}", text))

    return stages, anchors


samples = {
    "test_network": "test_network_baseline.log",
    "test_with_inetsim": "test_with_inetsim_baseline.log",
    "test_full_mode": "test_full_mode_baseline.log",
}

def report_group(title: str, samples_map: dict):
    total_stages = 0
    total_anchors = 0

    print(title)
    print(f"{'Sample':<25} {'Stages':>8} {'Anchors':>8}")
    print("-" * 45)

    for name, fname in samples_map.items():
        s, a = count_stages_and_anchors(LOG_DIR / fname)
        print(f"{name + (' (baseline)' if 'baseline' in fname else ''):<25} {str(s):>8} {str(a):>8}")
        if s:
            total_stages += s
        if a:
            total_anchors += a

    print(f"{'TOTAL':<25} {total_stages:>8} {total_anchors:>8}")
    print()

    return total_stages, total_anchors


if __name__ == "__main__":
    # Baseline group
    samples_base = {
        "test_network": "test_network_baseline.log",
        "test_with_inetsim": "test_with_inetsim_baseline.log",
        "test_full_mode": "test_full_mode_baseline.log",
    }
    base_stages, base_anchors = report_group("=== Baseline ===", samples_base)

    # Enhanced group
    samples_enh = {
        "test_with_inetsim": "test_with_inetsim_enhanced.log",
        "test_full_mode": "test_full_mode_enhanced.log",
    }
    enh_stages, enh_anchors = report_group("=== Enhanced ===", samples_enh)

    print("==> Ghi vao report:")
    print(f"  baseline_visible_stages = {base_stages}")
    print(f"  enhanced_visible_stages = {enh_stages}")
    print(f"  baseline_anchors        = {base_anchors}")
    print(f"  enhanced_anchors        = {enh_anchors}")

