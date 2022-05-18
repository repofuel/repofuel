import {
  Frequency,
  Period,
} from '../administration/components/__generated__/ActivityDashboardQuery.graphql';
import {eachDayOfInterval, eachMonthOfInterval, format, sub} from 'date-fns';
import {TimeUnit} from 'chart.js';

type DataPoints =
  | ReadonlyArray<{[key: string]: string | number}>
  | null
  | undefined;

export function fillPointsOnTimeLine(
  timeLine: Array<string>,
  dataPoints: DataPoints,
  y: string = 'y',
  x: string = 'x'
) {
  if (!dataPoints) {
    return new Array(timeLine.length).fill(0);
  }

  let lastIndex = dataPoints.length - 1;
  const res = Array(timeLine.length);

  for (let i = timeLine.length - 1; i >= 0; i--) {
    let time = timeLine[i];
    if (lastIndex >= 0) {
      const point = dataPoints[lastIndex];
      if (point[x] === time) {
        lastIndex--;
        res[i] = {x: time, y: point[y]};
        continue;
      } else if (point[x] > time) {
        lastIndex--;
        i++;
        continue;
      }
    }
    res[i] = {x: time, y: 0};
  }

  return res;
}

export function frequencyFromPeriod(period: Period): [Frequency, TimeUnit] {
  switch (period) {
    case 'MONTH':
    case 'WEEK':
      return ['DAILY', 'day'];
    case 'YEAR':
    case 'ALL_TIME':
      return ['MONTHLY', 'month'];
    default:
      throw Error('Unexpected period value: ' + period);
  }
}

export function generateTimeLine(period: Period, frequency: Frequency) {
  let interval: Interval = {start: 0, end: new Date()};
  if (period === 'MONTH') {
    interval.start = sub(interval.end, {months: 1});
  } else if (period === 'WEEK') {
    interval.start = sub(interval.end, {weeks: 1});
  } else if (period === 'YEAR') {
    interval.start = sub(interval.end, {years: 1});
  } else {
    throw Error('unexpected period: ' + period);
  }

  if (frequency === 'DAILY') {
    return eachDayOfInterval(interval).map((d) => format(d, 'yyyy-MM-dd'));
  }

  if (frequency === 'MONTHLY') {
    return eachMonthOfInterval(interval).map((d) => format(d, 'yyyy-MM'));
  }

  throw Error('unexpected frequency: ' + frequency);
}
