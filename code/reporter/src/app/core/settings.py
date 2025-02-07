"""
Copyright (c) 2022 Cyber Security & Privacy Foundation - All Rights Reserved
Unauthorized copying of this file, via any medium is strictly prohibited
Proprietary and confidential
Written by Cyber Security & Privacy Foundation
"""

from dataclasses import dataclass


@dataclass
class ScannerSetting:
    web_scan_timeout: int = 2700  # 45 minutes
    web_check_interval: int = 45  # 45 seconds

    network_scan_timeout: int = 2700  # 45 minutes
    network_check_interval: int = 45  # 45 seconds

    ip_range_timeout_per_ip: int = 1800  # 30 minutes
    ip_range_check_interval: int = 60  # 1 minute
    ip_range_max_timeout: int = 86400  # 24 hours


@dataclass
class AppSetting:
    """
    * Class to store app level settings
    """

    # Directories Reference
    config_dir: str
    app_data_dir: str
    logs_dir: str
    local_temp_dir: str
    user_dir: str
    main_config_path: str

    output_dir: str

    main_db_uri: str
    main_db_name: str
