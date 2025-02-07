from datetime import datetime


from app.db.models.base_model import BaseDocumentModel
from app.enums.main import ScanStatus
from app.utils import ip_utils
from zoneinfo import ZoneInfo


class Target(BaseDocumentModel):
    customer_username: str
    target_address: str
    target_type: str
    flag: int = 0
    scan_status: ScanStatus.ENUM
    created_at: datetime
    is_ds: bool = False
    scans: list | None = None
    scanner_ip: str | None = None
    scanner_username: str | None = None
    scan_started_time: datetime | None = None
    scan_completed_time: datetime | None = None
    overall_cvss_score: float | None = None
    cvss_score_by_host: dict | None = None
    failure_reason: int | None = None

    def is_ip_range(self):
        if ip_utils.is_valid_cidr_by_regex(self.target_address):
            return True

    def get_scan_started_time(self):
        if not self.scan_started_time:
            return ""

        # Assuming scan_started_time is stored in UTC, we ensure it's timezone-aware
        if self.scan_started_time.tzinfo is None:
            # If it's naive, assume it's in UTC
            self.scan_started_time = self.scan_started_time.replace(
                tzinfo=ZoneInfo("UTC")
            )
            print(f"Tzinfo is None: {self.scan_started_time}")
        print(f"Scan Started Time: {self.scan_started_time}")

        # Convert from UTC to the system's local timezone
        local_timezone = ZoneInfo("localtime")  # Uses the system's local timezone
        scan_started_time_local = self.scan_started_time.astimezone(local_timezone)

        return scan_started_time_local.strftime("%d-%m-%Y %I:%M%p")

    def get_scan_completed_time(self):
        if not self.scan_completed_time:
            return ""

        # Assuming scan_completed_time is stored in UTC, we ensure it's timezone-aware
        if self.scan_completed_time.tzinfo is None:
            # If it's naive, assume it's in UTC
            self.scan_completed_time = self.scan_completed_time.replace(
                tzinfo=ZoneInfo("UTC")
            )
            print(f"Tzinfo is None: {self.scan_completed_time}")
        print(f"Scan Completed Time: {self.scan_completed_time}")

        # Convert from UTC to the system's local timezone
        local_timezone = ZoneInfo("localtime")  # Uses the system's local timezone
        scan_completed_time_local = self.scan_completed_time.astimezone(local_timezone)

        return scan_completed_time_local.strftime("%d-%m-%Y %I:%M%p")
