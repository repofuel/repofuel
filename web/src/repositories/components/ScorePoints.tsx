import React from 'react';

import {ReactComponent as ScoreLevel0} from '../svg/scores/0.svg';
import {ReactComponent as ScoreLevel1} from '../svg/scores/1.svg';
import {ReactComponent as ScoreLevel2} from '../svg/scores/2.svg';
import {ReactComponent as ScoreLevel3} from '../svg/scores/3.svg';
import {ReactComponent as ScoreLevel4} from '../svg/scores/4.svg';
import {ReactComponent as ScoreLevel5} from '../svg/scores/5.svg';
import {ReactComponent as ScoreLevel6} from '../svg/scores/6.svg';
import {ReactComponent as ScoreLevel7} from '../svg/scores/7.svg';
import {ReactComponent as ScoreLevel8} from '../svg/scores/8.svg';
import {ReactComponent as ScoreLevel9} from '../svg/scores/9.svg';
import {ReactComponent as ScoreLevel10} from '../svg/scores/10.svg';

interface ScorePointsProps extends React.SVGAttributes<SVGElement> {
  level: number;
}

export const ScorePoints: React.FC<ScorePointsProps> = ({level, ...props}) => {
  switch (level) {
    case 0:
      return <ScoreLevel0 {...props} />;
    case 1:
      return <ScoreLevel1 {...props} />;
    case 2:
      return <ScoreLevel2 {...props} />;
    case 3:
      return <ScoreLevel3 {...props} />;
    case 4:
      return <ScoreLevel4 {...props} />;
    case 5:
      return <ScoreLevel5 {...props} />;
    case 6:
      return <ScoreLevel6 {...props} />;
    case 7:
      return <ScoreLevel7 {...props} />;
    case 8:
      return <ScoreLevel8 {...props} />;
    case 9:
      return <ScoreLevel9 {...props} />;
    case 10:
      return <ScoreLevel10 {...props} />;
    default:
      return <ScoreLevel0 {...props} />; //todo: use placeholder
  }
};

export default ScorePoints;
