import unittest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import oracle_probe


class OracleProbeSummaryTests(unittest.TestCase):
    def test_summary_contains_only_status_metadata(self):
        results = [
            oracle_probe.ProbeResult(
                name="equipment_rank",
                status="pass",
                detail="private detail",
                output="/private/path/probe.json",
            ),
            oracle_probe.ProbeResult(
                name="dino_cryopod_location",
                status="error",
                detail="private traceback",
                output="/private/path/probe.json",
            ),
        ]

        got = oracle_probe.summary_rows(results)

        self.assertEqual(
            got,
            [
                {
                    "name": "equipment_rank",
                    "status": "pass",
                    "detail": "private detail",
                },
                {
                    "name": "dino_cryopod_location",
                    "status": "error",
                    "detail": "private traceback",
                },
            ],
        )

    def test_default_output_path_uses_oracle_output(self):
        got = oracle_probe.default_output_path(
            oracle_probe.Path("/repo"),
            "equipment_rank",
        )

        self.assertEqual(got, oracle_probe.Path("/repo/.oracle/output/oracle-probe-equipment-rank.json"))


if __name__ == "__main__":
    unittest.main()
