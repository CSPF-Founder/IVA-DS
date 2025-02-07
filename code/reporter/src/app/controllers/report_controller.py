from bson.objectid import ObjectId

from app import core_app
from app.db.connections.main_db import MainDatabase
from app.db.crud.target import CrudTarget

from app.controllers.web.reporter import Reporter as WebReporter
from app.controllers.web.reporter import DSReporter as WebDSReporter

from app.controllers.network.reporter import Reporter as NetworkReporter
from app.controllers.network.reporter import DSReporter as NetworkDSReporter


def run_reporter(target_id: ObjectId):
    with MainDatabase() as db:
        crud_target = CrudTarget(db)
        target = crud_target.find_by_id(target_id=target_id)

    if not target:
        raise Exception("Unable to find the target with object id {}".format(target_id))

    core_app.logger.info(
        "Running Reporter for :"
        + target.target_address
        + " | Customer:"
        + target.customer_username
    )

    if target.target_type == "url":
        reporter = WebDSReporter(target=target) if target.is_ds else WebReporter(target=target)
    elif target.target_type in ["ip", "ip_range"]:
        reporter = NetworkDSReporter(target=target) if target.is_ds else NetworkReporter(target=target)
    else:
        raise Exception("Invalid target category")

    reporter.run()