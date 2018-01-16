import React, { Component } from 'react';
import Notifications from 'react-notify-toast';
import PagedrawGeneratedPage from './pagedraw/app'
import $ from 'jquery';
import {emptyGitInfo, fetchGitInfo, handlePull, handlePush} from './Api'

class App extends Component {
  constructor() {
      super();
      this.state = {
          filename: "",
          content: "",
          searchResultList: [],
          gitInfo: emptyGitInfo(),
      };

      // Kick off populating git info
      this.populateGitInfo();
  }

  componentWillReceiveProps() {
      // Kick off populating git info
      this.populateGitInfo();
  }

  populateGitInfo() {
      fetchGitInfo((newInfo) => {
          this.setState({
              gitInfo: newInfo,
          })
      })
  }

  handleFilenameChange(e) {
       this.setState({
           filename: e.target.value,
      })
  }

  handleContentChange(e) {
      this.setState({
          content: e.target.value,
      })
  }

  handleSaveNote() {
      $.post("/api/1/commit", {
          filename: this.state.filename,
          content: this.state.content,
      }).done(() => {
          // TODO: Something nicer on the refresh side?
          window.location = "/"
      })
  }

  handleSearchQueryChange(e) {
      // TODO: Debounce?
      $.post("/api/1/search", {
          query: e.target.value,
      }).done((data) => {
          this.setState({
              searchResultList: JSON.parse(data),
          })
      })
  }

  handleExpand(e) {
      let cur = e.target.parentElement;
      let path = "";

      // Walk up the tree to generate the rest of the path
      let last = null;
      while (cur.className.indexOf("folderlist") === 0) {
          // This will handle the base case
          let possibleName = cur.parentElement.parentElement.children[0].children[0].children[0];
          if (possibleName.className.indexOf("folderlistelement-name") === 0) {
              path = "/" + possibleName.innerText + path;
              last = possibleName;
          } else if (cur.children[0].className.indexOf("folderlist-folderlist") === 0) {
              // This will extract names out of containing folderlists
              let child = cur;
              while (child.className.indexOf("folderlistelement-name") !== 0) {
                  child = child.children[0];
              }
              if (last !== child) {
                  // This prevents us from double counting clicked-on containing folders in folderlists
                  path = "/" + child.innerText + path;
              }
          }

          cur = cur.parentElement
      }
      console.log("Clicked on ", path);
      let pathParts = path.split("/");
      // TODO: Remove this horrible hack and instead let there be two handle functions separately
      // for search results / root folder list
      let curNode = {contents: this.props.rootFolderList.concat(this.state.searchResultList)};
      for (let i = 1; i < pathParts.length; i++) {
          let target = pathParts[i];
          let found = false;
          for (let j = 0; j < curNode.contents.length; j++) {
              if (curNode.contents[j].name === target) {
                  found = true;
                  curNode = curNode.contents[j];
                  break
              }
          }
          if (!found) {
              throw Error("Unable to traverse back down tree")
          }
      }
      if (curNode.state === "file") {
          // Open the file by navigating to the edit page for it.
          window.location = "/edit/" + curNode.id
      } else if (curNode.state === "expanded" || curNode.state === "expandedEmpty") {
          curNode.state = "collapsed"
      } else if (curNode.state === "collapsed") {
          if (curNode.contents.length > 0) {
              curNode.state = "expanded"
          } else {
              curNode.state = "expandedEmpty"
          }
      }
      this.setState(this.state)
  }

  render() {
    return <div>
        <PagedrawGeneratedPage
          rootFolderList={this.props.rootFolderList}
          filename={this.state.filename}
          content={this.state.content}
          searchResultList={this.state.searchResultList}
          handleSearchQueryChange={this.handleSearchQueryChange.bind(this)}
          handleFilenameChange={this.handleFilenameChange.bind(this)}
          handleContentChange={this.handleContentChange.bind(this)}
          handleSaveNote={this.handleSaveNote.bind(this)}
          handleExpand={this.handleExpand.bind(this)}
          handlePull={handlePull}
          handlePush={handlePush}
          lastCommit={this.state.gitInfo.lastCommit}
          lastPull={this.state.gitInfo.lastPull}
          remoteAheadBy={this.state.gitInfo.remoteAheadBy}
          localAheadBy={this.state.gitInfo.localAheadBy}
        />
        <Notifications />
    </div>;

  }
}

export default App;
