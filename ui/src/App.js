import React, { Component } from 'react';
import PagedrawGeneratedPage from './pagedraw/app'
import $ from 'jquery';

class App extends Component {
  constructor(props) {
      super();
      this.state = {
          filename: "",
          content: "",
          rootFolderList: props.rootFolderList,
      }
  }

  handleFilenameChange(e) {
       this.setState({
          filename: e.target.value,
          content: this.state.content,
      })
  }

  handleContentChange(e) {
      this.setState({
          filename: this.state.filename,
          content: e.target.value,
      })
  }

  handleSaveNote() {
      $.post("/api/1/save", {
          filename: this.state.filename,
          content: this.state.content,
      }).done(() => {
          // TODO: Something nicer on the refresh side?
          window.location = "/"
      })
  }

  handlePull() {
      $.get("/api/1/pull", () => {
          window.location = "/"
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
      let curNode = {contents: this.state.rootFolderList};
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
    return (
        <PagedrawGeneratedPage
          rootFolderList={this.state.rootFolderList}
          filename={this.state.filename}
          content={this.state.content}
          handleFilenameChange={this.handleFilenameChange.bind(this)}
          handleContentChange={this.handleContentChange.bind(this)}
          handleSaveNote={this.handleSaveNote.bind(this)}
          handleExpand={this.handleExpand.bind(this)}
          handlePull={this.handlePull}
        />
    );
  }
}

export default App;
