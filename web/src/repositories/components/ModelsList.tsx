import React, {useState} from 'react';
import {Model, ModelDataPoints, ModelReport} from '../types';
import {
  CollapsibleList,
  List,
  ListItem,
  ListItemMeta,
  ListItemText,
} from '@rmwc/list';
import {Tooltip} from '@rmwc/tooltip';
import {format, formatDistanceToNow} from 'date-fns';
import {FontAwesomeIcon} from '@fortawesome/react-fontawesome';
import {faChevronDown, faChevronRight} from '@fortawesome/free-solid-svg-icons';
import ReactJson from 'react-json-view';

interface ModelsListProps {
  models?: Model[];
}

export const ModelsList: React.FC<ModelsListProps> = ({models = []}) => {
  return (
    <List className="divided-collapsible-list">
      {models.map((model, i) => (
        <ModelsListItem key={model.id} model={model} />
      ))}
    </List>
  );
};

interface ModelsListItemProps {
  model: Model;
}

const ModelsListItem: React.FC<ModelsListItemProps> = ({model}) => {
  const [isOpen, setOpen] = useState(false);

  const isExpandable = model.report;
  return (
    <CollapsibleList
      handle={
        <ListItem onClick={() => isExpandable && setOpen(!isOpen)}>
          <ListItemText>
            Version {model.version}
            <Tooltip
              content={format(
                new Date(model.created_at),
                "'Built on' LLL M, y 'at' h:m a"
              )}
              showArrow
              align="right">
              <span className="process-time">
                {' '}
                built{' '}
                {formatDistanceToNow(new Date(model.created_at), {
                  addSuffix: true,
                })}
              </span>
            </Tooltip>
          </ListItemText>
          <ListItemMeta>
            {isExpandable && (
              <FontAwesomeIcon icon={isOpen ? faChevronDown : faChevronRight} />
            )}
          </ListItemMeta>
        </ListItem>
      }>
      {isOpen && isExpandable && (
        <ModelsListItemDetails
          data={model.data}
          report={model.report}
          medians={model.medians}
        />
      )}
    </CollapsibleList>
  );
};

interface ModelsListItemDetailsProps {
  report: ModelReport;
  data: ModelDataPoints;
  medians: any;
}

const ModelsListItemDetails: React.FC<ModelsListItemDetailsProps> = ({
  data,
  report,
  medians,
}) => {
  return (
    <ReactJson
      name="model"
      displayDataTypes={false}
      displayObjectSize={false}
      style={{marginLeft: '1em'}}
      src={{medians, data, report}}
    />
  );
};
