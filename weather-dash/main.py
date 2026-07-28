import time
import datetime as dt


todays_date: str = str(time.localtime())

banner: str = "Good {time_of_day}"

print(f'{banner} /n {todays_date}')
