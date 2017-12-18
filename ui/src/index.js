import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import registerServiceWorker from './registerServiceWorker';

// Read in the files.

ReactDOM.render(<App rootFolderList={[
              {
                  name: "folder1",
                  state: "collapsed",
              },
              {
                  name: "file1",
                  state: "file",
              },
              {
                  name: "folder2Empty",
                  state: "expandedEmpty",
              },
              {
                  name: "folder3",
                  state: "expanded",
                  contents: [
                      {
                          name: "subfile",
                          state: "file"
                      },
                      {
                          name: "subfile2",
                          state: "file"
                      },
                      {
                          name: "folder4",
                          state: "expanded",
                          contents: [
                              {
                                  name: "subsubfile1",
                                  state: "file",
                              }
                          ]
                      }
                  ]
              },
]}/>, document.getElementById('root'));
registerServiceWorker();
