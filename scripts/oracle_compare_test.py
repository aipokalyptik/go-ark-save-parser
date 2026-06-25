import unittest

import oracle_compare


class OracleCompareParserTests(unittest.TestCase):
    def test_parse_go_heatmap_summary(self):
        got = oracle_compare.parse_go_heatmap_summary(
            "cells=12 total=34 max=5 faults=2 wrote=.oracle/output/dino-heatmap.json\n"
        )

        self.assertEqual(
            got,
            {
                "nonzero_cells": 12,
                "total": 34,
                "max": 5,
                "faults": 2,
            },
        )


if __name__ == "__main__":
    unittest.main()
