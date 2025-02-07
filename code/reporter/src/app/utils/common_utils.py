"""
Copyright (c) 2022 Cyber Security & Privacy Foundation - All Rights Reserved
Unauthorized copying of this file, via any medium is strictly prohibited
Proprietary and confidential
Written by Cyber Security & Privacy Foundation
"""


from typing import Any


def smart_str(input_text: Any) -> str:
    """
    * Proper handling for string
    * Converts any input like unicode,numbers into string value

    Args:
        input_text (_type_): _description_

    Returns:
        _type_: _description_
    """
    if not input_text:
        return ""

    if isinstance(input_text, str):
        return input_text
    if isinstance(input_text, (bytearray, bytes)):
        return str(input_text, "utf-8")
    if isinstance(input_text, (int, float)):
        return str(input_text)

    return str(input_text, "utf-8")
