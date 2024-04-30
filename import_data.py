import csv
from influxdb_client import InfluxDBClient, Point, WritePrecision, WriteOptions
from influxdb_client.client.write_api import SYNCHRONOUS
from tqdm import tqdm

from datetime import datetime

url = 'https://influxdb2.qcloud.ry.rs'  
token = '5XmzU06hcSVsO4b2gC5R5_blpHpY-JQyG_V70map_RTmYQQQKjWCOL45M0umqSuXXyb6pfkkabaBwhyz0YRRDw=='  
org = 'MyOrg'  
bucket = 'MyBucket'  

with InfluxDBClient(url=url, token=token, org=org) as client:
    write_api = client.write_api(write_options=SYNCHRONOUS)
    csv_file_path = 'data.csv'
    with open(csv_file_path, 'r') as file:
        reader = csv.DictReader(file)
        all_rows = list(reader)
        batch_size = 100

        batches = [all_rows[i:i+batch_size] for i in range(0, len(all_rows), batch_size)]

        for batch in tqdm(batches):
            points = []
            for row in batch:
                point = Point("ahr999") \
                    .field("value", float(row['value'])) \
                    .field("avg", float(row['avg'])) \
                    .field("ahr999", float(row['ahr999'])) \
                    .time(datetime.fromisoformat(row['time'][:-1]), WritePrecision.NS)
                points.append(point)
            write_api.write(bucket=bucket, org=org, record=points)
