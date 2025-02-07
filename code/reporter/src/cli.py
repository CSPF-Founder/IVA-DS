import argparse
from bson.objectid import ObjectId

# from app.controllers.scan_controller import run_scan
from app.controllers.report_controller import run_reporter


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-t", "--target-id", dest="target_id", required=True)

    args = parser.parse_args()
    if ObjectId.is_valid(args.target_id):
        target_id = ObjectId(args.target_id)
        run_reporter(target_id)

    else:
        print("Invalid object id received")
        exit(1)


if __name__ == "__main__":
    main()
