import * as React from "react";
import * as ReactDOM from "react-dom";
import { observer } from "mobx-react";
import { Store } from "../../store";
import { LoginPopup } from "../popups/login";
import { NotificationsPopup } from "../popups/notifications";
import { Icon, IconGroup } from "semantic-ui-react"

export interface NotificationsIconProps {
  store: Store,
}

@observer
export class NotificationsIcon extends React.Component<NotificationsIconProps, {}> {
  render() {
    const store = this.props.store

    let unread = false;
    for (let n of store.notifications) {
      if (n.read === null) {
        unread = true;
        break;
      }
    }

    return (<IconGroup size="large">
      <Icon name="alarm" />
      {unread ? <Icon corner color="red" name="circle" /> : null}
    </IconGroup>)
  }
}

