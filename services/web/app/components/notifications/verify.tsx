import * as React from "react";
import { Store } from "../../store"
import { Notification } from "../../models"
import { Item, Button, Icon } from "semantic-ui-react";
import { ReadIcon } from "../icons/read";

export interface VerifyEmailNotificationProps {
  store: Store;
  notification: Notification;
}

export class VerifyEmailNotification extends React.Component<VerifyEmailNotificationProps, {}> {

  render() {
    const n = this.props.notification;
    return (
      <Item>
      <Item.Content verticalAlign='middle'>
        <Item.Header><ReadIcon notification={n} />Email verification</Item.Header>
        <Item.Description>Please check your email for a message from us.</Item.Description>
        <Item.Extra>
          <Button floated='right' icon="close" size="tiny" basic />
          <Button floated='right' size="tiny">
            Resend
          </Button>
        </Item.Extra>
      </Item.Content>
    </Item>
    )
  }

}

