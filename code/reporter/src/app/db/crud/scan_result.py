from typing import List
from bson.objectid import ObjectId

from app.enums.main import AlertStatus
from app.db.crud.base import CrudBase
from app.db.models.target import Target


class CrudScanResult(CrudBase):
    COLLECTION_NAME = "scan_results"

    def __init__(self, db_session) -> None:
        super().__init__(db_session, self.COLLECTION_NAME)

    def add(self, records: list) -> int:
        if not records:
            return 0
        inserted_result = self.collection.insert_many(records)
        if not inserted_result:
            return 0

        return len(inserted_result.inserted_ids)

    def exists(self, filters: dict) -> bool:
        if not filters:
            return False

        result = self.collection.find_one(filters)
        if not result:
            return False

        return True

    def get_list_by_target(self, target: Target) -> List:
        result_cursor = self.collection.find({"target_id": target.id}).sort("severity")

        entries = []
        if not result_cursor:
            return []
        for doc in result_cursor:
            entries.append(doc)

        return entries

    def get_list_by_target_id(self, target_id):
        if not target_id:
            raise Exception("Empty Mongodb object id given")

        if not ObjectId.is_valid(target_id):
            raise Exception("Invalid Mongodb object id given")

        if not isinstance(target_id, ObjectId):
            target_id = ObjectId(target_id)

        result_cursor = self.collection.find({"target_id": target_id}).sort("severity")

        entries = []
        if not result_cursor:
            return
        for doc in result_cursor:
            entries.append(doc)

        return entries


class CrudDSScanResult(CrudBase):
    COLLECTION_NAME = "scan_results_"

    def __init__(self, db_session, target_id) -> None:
        super().__init__(db_session, self.COLLECTION_NAME + str(target_id))

    def get_list_by_target(self, target: Target):
        result_cursor = self.collection.find({"target_id": target.id}).sort("severity")

        entries = []
        if not result_cursor:
            return
        for doc in result_cursor:
            entries.append(doc)

        return entries

    def alerts_by_status(self, alert_status: int) -> list:
        if alert_status not in [
            AlertStatus.ENUM.UNFIXED.value,
            AlertStatus.ENUM.FIXED.value,
        ]:
            raise Exception("Invalid alert type given")

        result_cursor = self.collection.find({"alert_status": alert_status}).sort(
            "severity"
        )

        entries = []
        if not result_cursor:
            return []

        for doc in result_cursor:
            entries.append(doc)

        return entries
