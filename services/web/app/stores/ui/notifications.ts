import * as React from "react";
import { observable, autorun } from "mobx";
import "whatwg-fetch"

/**
 * A data store for tracking interface notifications (i.e., the nag at the
 * top of a page.)
 */
export class NotificationStore {

  notifications: Notification[] = [];

  push = (n: Notification) => {

  }
}

export class Notification {

  @observable dismissed: boolean = false;

  view: any;

  constructor(view: any) {
    this.view = view;
  }
}

