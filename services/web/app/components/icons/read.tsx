import * as React from "react";
import { observer } from "mobx-react";
import { Store } from "../../store"
import { Notification } from "../../models"
import { Icon } from "semantic-ui-react";

export interface ReadIconProps {
  notification: Notification;
}

@observer
export class ReadIcon extends React.Component<ReadIconProps, {}> {

  render() {
    const n = this.props.notification;
    return n.read === null ? (
      <Icon name="circle" color="grey" size="mini" style={{verticalAlign:"0.5em"}}/> 
    ) : null;
  }

}


