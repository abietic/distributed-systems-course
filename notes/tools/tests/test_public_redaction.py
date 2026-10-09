"""Regression checks for masking without breaking event references."""

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

TOOLS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOLS))
from public_redaction import PATTERNS, PublicMasker


class PublicRedactionTests(unittest.TestCase):
    def test_aliases_preserve_call_and_result_references(self):
        first = "11111111-1111-4111-8111-111111111111"
        second = "22222222-2222-4222-8222-222222222222"
        tool_id = "toolu_" + "A" * 24
        events = [
            {"event_id": first, "sequence_num": 1, "created_at": "2026-08-20T02:16:00Z",
             "payload": {"tool_use": {"id": tool_id, "input": "x += 1"}}},
            {"event_id": second, "sequence_num": 2, "created_at": "2026-08-20T02:17:00Z",
             "payload": {"tool_use_id": tool_id, "parent_uuid": first, "result": "x = 1"}},
        ]
        readable = f"event {first}: call {tool_id}; result {second}"
        masker = PublicMasker()
        masker.prime(events, readable)
        masked = masker.deep(events)
        self.assertNotEqual(masked[0]["event_id"], masked[1]["event_id"])
        self.assertEqual(masked[0]["event_id"], masked[1]["payload"]["parent_uuid"])
        call = masked[0]["payload"]["tool_use"]["id"]
        self.assertEqual(call, masked[1]["payload"]["tool_use_id"])
        self.assertIn(call, masker.text(readable))
        self.assertEqual([e["sequence_num"] for e in masked], [1, 2])
        self.assertEqual(masked[0]["created_at"], events[0]["created_at"])
        self.assertEqual(masked[0]["payload"]["tool_use"]["input"], "x += 1")
        for original in (first, second, tool_id):
            self.assertNotIn(original, json.dumps(masked))

    def test_credentials_images_paths_and_account_fields(self):
        original = {
            "session_id": "session-example-opaque-id", "account_id": "account-example-opaque-id",
            "device_name": "Example Laptop", "trace_context": "00-" + "a" * 32 + "-" + "b" * 16 + "-01",
            "signature": "short-signature-secret", "access_token": "short-token-secret",
            "image": {"type": "base64", "data": "A" * 240, "media_type": "image/png"},
            "text": "Read /Users/example/Code/course/demo.go; contact reader@example.invalid at 10.2.3.4",
        }
        masker = PublicMasker()
        masker.prime(original)
        masked = masker.deep(original)
        for key in ("session_id", "account_id", "device_name", "trace_context", "signature", "access_token"):
            self.assertNotEqual(masked[key], original[key])
        self.assertEqual(masked["image"]["media_type"], "image/png")
        text = json.dumps(masked, ensure_ascii=False)
        self.assertFalse(any(pattern.search(text) for pattern in PATTERNS.values()))

    def test_export_is_deterministic(self):
        original = {"session_id": "session-example-opaque-id", "text": "/home/example/course.md"}
        a, b = PublicMasker(), PublicMasker()
        a.prime(original)
        b.prime(original)
        self.assertEqual(a.deep(original), b.deep(original))

    def test_checker_scans_raw_and_decodes_escaped_json(self):
        with tempfile.TemporaryDirectory() as tmp:
            subprocess.run(["git", "init", "-q", tmp], check=True)
            raw = Path(tmp) / "notes" / "raw" / "events.jsonl"
            raw.parent.mkdir(parents=True)
            raw.write_text('{"content":"\\u0072eader@example.invalid"}\n')
            run = subprocess.run([sys.executable, str(TOOLS / "check_redaction.py")], cwd=tmp,
                                 capture_output=True, text=True)
            self.assertEqual(run.returncode, 1)
            self.assertIn("email", run.stdout)
            self.assertNotIn("reader@example.invalid", run.stdout)
            raw.write_text('{"content":"〔邮箱00001〕"}\n')
            run = subprocess.run([sys.executable, str(TOOLS / "check_redaction.py")], cwd=tmp,
                                 capture_output=True, text=True)
            self.assertEqual(run.returncode, 0)

    def test_checker_rejects_non_uuid_account_id_and_short_signature(self):
        from check_redaction import metadata_findings
        raw = {"payload": {"account_id": "opaque-account-reference", "signature": "short-secret"}}
        self.assertEqual(metadata_findings(raw), {"unmasked metadata ID", "unmasked credential/signature field"})
        masker = PublicMasker()
        masker.prime(raw)
        self.assertEqual(metadata_findings(masker.deep(raw)), set())


if __name__ == "__main__":
    unittest.main()
