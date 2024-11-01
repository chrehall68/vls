module Main where

import Scanner (alexScanTokens)

main :: IO ()
main = do
  s <- getContents
  print (alexScanTokens s)
