'use strict';
/*
 * Copyright 2019 Hayo van Loon
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

function doHttpGet(url, callback, errorCb) {
  sendHttpReq("GET", url, null, null, null, callback, errorCb);
}

function doHttpGetAuth(url, idToken, callback, errorCb) {
  sendHttpReq("GET", url, null, null, {'Authorization': 'Bearer ' + idToken}, callback, errorCb);
}

function doHttpPostForm(url, body, callback, errorCb) {
  sendHttpReq("POST", url, body, "application/x-www-form-urlencoded", null, callback, errorCb);
}

function doHttpPostJson(url, body, callback, errorCb) {
  sendHttpReq("POST", url, body, "application/json", null, callback,
      errorCb);
}

function sendHttpReq(method, url, body, contentType, headers, callback, errorCb) {
  let httpReq = new XMLHttpRequest();
  httpReq.onreadystatechange = function () {
    if (httpReq.readyState === 4 && httpReq.status === 200) {
      callback(httpReq.responseText);
    } else if (typeof (errorCb) === typeof (Function) &&
        httpReq.readyState === 4 &&
        httpReq.status >= 400) {
      errorCb(httpReq);
    }
  };
  httpReq.open(method, url, true);
  if (!!headers) {
    for (let k in headers) {
      if (headers.hasOwnProperty(k)) {
        httpReq.setRequestHeader(k, headers[k]);
      }
    }
  }
  if (!!body) {
    httpReq.setRequestHeader("Content-type", contentType);
  }
  httpReq.send(!!body ? body : null);
}


function pollUntil(url, isDone, timeout, freq, doneEvent) {
  isDone = isDone || function (e) { return true; };
  timeout = timeout || 60000;
  freq = freq || 2;
  doneEvent = doneEvent || new CustomEvent("pollDone", {});

  let delay = 1000 / freq;

  if (timeout > 0) {
    var callback = function (resp) {
      if (isDone(resp)) document.dispatchEvent(doneEvent);
      else {
        window.setTimeout(function () { pollUntil(url, isDone, timeout - delay); }, delay);
      }
    };
    var errorCallback = function (e) {
      document.dispatchEvent(new CustomEvent("pollError", {'detail': e}));
    };
    doHttpGet(url, callback, errorCallback);
    document.dispatchEvent(new CustomEvent("pollSent", {}));
  } else {
    document.dispatchEvent(new CustomEvent("pollTimeout", {}));
  }
}

