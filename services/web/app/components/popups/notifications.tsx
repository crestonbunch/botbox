import * as React from "react";
import { observer } from "mobx-react";
import { action } from "mobx";
import { Item, Popup } from "semantic-ui-react";
import { Api } from "../../api";
import { Store } from "../../store";
import { Notification } from "../../models";
import { VerifyEmailNotification } from "../notifications/verify";

export interface NotificationsPopupPros {
  trigger: JSX.Element;
  store: Store;
}

@observer
export class NotificationsPopup extends React.Component<NotificationsPopupPros, {}> {

  @action onOpen() {
    const store = this.props.store;
    setTimeout(action("readTimeout", () => {
      let unread = store.notifications.filter((n: Notification) => {
        return n.read == null;
      });
      if (unread.length > 0) {
        Api.read(store.session.secret, unread).then(function () {
            unread.map((v: Notification) => { v.read = new Date() });
        });
      }
    }), 1000);
  }

  render() {
    const store = this.props.store;

    let list: any[] = [];
    for (let n of store.notifications) {
      switch (n.type) {
        case "verify":
          list.push(<VerifyEmailNotification store={store} notification={n} />);
          break;
      }
    }

    let group = <Item.Group divided>{list}</Item.Group>

    return (<Popup
      trigger={this.props.trigger}
      onOpen={this.onOpen.bind(this)}
      content={group}
      on='click'
      positioning='bottom center'
      wide='very'
    />)
  }

}
