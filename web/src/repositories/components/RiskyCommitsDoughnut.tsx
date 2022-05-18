import React, {useEffect, useRef} from 'react';
import {Chart} from 'chart.js';

export const RiskyCommitsDoughnut: React.FC<any> = ({
  commitNum,
  riskyCommitNum,
  className,
}) => {
  let doughnut: any = useRef(null);
  useEffect(() => {
    if (doughnut == null) {
      return;
    }

    let ctxChart = doughnut.current.getContext('2d');

    new Chart(ctxChart, {
      type: 'doughnut',
      data: {
        labels: ['Risky Commits', 'Commits'],
        datasets: [
          {
            backgroundColor: [
              'rgba(255, 100, 100, 0.3)',
              'rgba(0, 220, 0, 0.3)',
            ],
            data: [riskyCommitNum, commitNum - riskyCommitNum],
          },
        ],
      },
      options: {
        responsive: true,
        legend: {
          display: false,
        },
        tooltips: {
          enabled: false,
        },
        cutoutPercentage: 70,
        hover: {mode: undefined},
      },

      plugins: [
        {
          beforeDraw: function (chart: Chart) {
            // Used to place text in the middle of the doughnut chart
            // Check if the passed variable is null or not to prevent null values
            // and potential errors
            if (chart == null || chart.ctx == null) {
              return;
            }
            let width = chart.width || 0,
              height = chart.height || 0,
              ctx = chart.ctx;

            ctx.restore();

            let fontSize = (70 / 114).toFixed(2); // 70/114 looks good for the size

            ctx.font = fontSize + 'em sans-serif';
            ctx.textBaseline = 'middle';

            let riskyCmtPct =
              riskyCommitNum && commitNum
                ? (riskyCommitNum / commitNum) * 100
                : 0;
            let text = riskyCmtPct.toFixed(0) + '%',
              textX = Math.round((width - ctx.measureText(text).width) / 2),
              textY = height / 2;

            ctx.fillText(text, textX, textY);
            ctx.save();
          },
        },
      ],
    });
  });

  return (
    <span className={className}>
      <canvas ref={doughnut} />
    </span>
  );
};
