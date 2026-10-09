"""Shared masking and detection rules for public learning records."""

import re


PATTERNS = {
    "UUID": re.compile(r"\b[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}\b", re.I),
    "provider ID": re.compile(r"\b(?:sess|msg|toolu)_[A-Za-z0-9_-]{16,}\b"),
    "local path": re.compile(r"/(?:Users|home|tmp|var|Applications|Volumes|workspace|mnt|opt|private)/[^\s`\"'<>\[\](),;\\]+"),
    "email": re.compile(r"(?<![\w.+-])[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}"),
    "private IP": re.compile(r"(?<!\d)(?:10\.(?:\d{1,3}\.){2}\d{1,3}|192\.168\.(?:\d{1,3}\.)\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.(?:\d{1,3}\.)\d{1,3})(?!\d)"),
    "credential": re.compile(r"\b(?:(?:gh[pousr]_|github_pat_)[A-Za-z0-9_]{16,}|(?:AKIA|ASIA)[A-Z0-9]{16}|sk-(?:ant-)?[A-Za-z0-9_-]{20,})\b"),
    "bearer token": re.compile(r"\bBearer\s+[A-Za-z0-9._~+/-]{20,}", re.I),
    "private key": re.compile(r"-----BEGIN (?:[A-Z]+ )?PRIVATE KEY-----.*?-----END (?:[A-Z]+ )?PRIVATE KEY-----", re.S),
    "opaque data": re.compile(r"[A-Za-z0-9+/=]{200,}"),
}
IDENTIFIER_FIELDS = {
    "event_id", "uuid", "session_id", "sent_by_account_id", "account_id",
    "organization_id", "org_id", "device_id", "trace_context", "x-session-uuid",
    "request_id", "file_uuid", "fileuuid", "logical_parent_uuid",
    "summarizes_uuid", "user_message_uuid", "hook_id",
    "username", "user_name", "device_name", "organization_name", "org_name", "account_name",
}
SECRET_FIELDS = {"signature", "encrypted_content", "password", "passwd",
                 "api_key", "apikey", "access_token", "refresh_token", "authorization"}


def strings(value):
    if isinstance(value, dict):
        for key, item in value.items():
            yield key
            yield from strings(item)
    elif isinstance(value, list):
        for item in value:
            yield from strings(item)
    elif isinstance(value, str):
        yield value


class PublicMasker:
    """Stable aliases within an export keep references without publishing originals."""

    def __init__(self):
        self.identifiers = {}
        self.paths = {}
        self.emails = {}
        self.ips = {}

    @staticmethod
    def alias(table, value, label):
        if value not in table:
            table[value] = f"〔{label}{len(table) + 1:05d}〕"
        return table[value]

    def prime(self, *sources):
        ids, paths, emails, ips = set(), set(), set(), set()
        for source in sources:
            for text in strings(source):
                ids.update(PATTERNS["UUID"].findall(text))
                ids.update(PATTERNS["provider ID"].findall(text))
                paths.update(PATTERNS["local path"].findall(text))
                emails.update(PATTERNS["email"].findall(text))
                ips.update(PATTERNS["private IP"].findall(text))
            self._collect_ids(source, ids)
        for table, values, label in ((self.identifiers, ids, "标识"),
                                      (self.paths, paths, "路径"),
                                      (self.emails, emails, "邮箱"),
                                      (self.ips, ips, "内网地址")):
            for value in sorted(values):
                self.alias(table, value, label)
        self.id_re = re.compile("|".join(re.escape(x) for x in sorted(self.identifiers, key=lambda x: (-len(x), x)))) if ids else None

    def _collect_ids(self, value, ids):
        if isinstance(value, dict):
            for key, item in value.items():
                if key.lower() in IDENTIFIER_FIELDS and isinstance(item, str) and item and not item.startswith("〔"):
                    ids.add(item)
                self._collect_ids(item, ids)
        elif isinstance(value, list):
            for item in value:
                self._collect_ids(item, ids)

    def text(self, text):
        for name in ("private key", "credential", "bearer token", "opaque data"):
            text = PATTERNS[name].sub(f"〔已脱敏：{name}〕", text)
        text = PATTERNS["local path"].sub(lambda m: self.alias(self.paths, m.group(), "路径"), text)
        text = PATTERNS["email"].sub(lambda m: self.alias(self.emails, m.group(), "邮箱"), text)
        text = PATTERNS["private IP"].sub(lambda m: self.alias(self.ips, m.group(), "内网地址"), text)
        if getattr(self, "id_re", None):
            text = self.id_re.sub(lambda m: self.identifiers[m.group()], text)
        for name in ("UUID", "provider ID"):
            text = PATTERNS[name].sub(lambda m: self.alias(self.identifiers, m.group(), "标识"), text)
        return text

    def deep(self, value):
        if isinstance(value, dict):
            return {self.text(key): ("〔已脱敏：凭据或签名〕" if key.lower() in SECRET_FIELDS and item
                                    else self.deep(item)) for key, item in value.items()}
        if isinstance(value, list):
            return [self.deep(item) for item in value]
        return self.text(value) if isinstance(value, str) else value
