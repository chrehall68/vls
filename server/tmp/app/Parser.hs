module Parser where

import Control.Monad
import Tokens (Token (..))

data VerilogModuleItem = VerilogModuleItem
  { itemIdentifier :: String
  }

data VerilogModule = VerilogModule
  { moduleName :: String,
    moduleParameters :: [VerilogModuleItem],
    moduleItems :: [VerilogModuleItem]
  }

-- returns (module name, rest of tokens)
parseParams :: [Token] -> ([VerilogModuleItem], [Token])
parseParams (IDENTIFIER name, COMMA : rest) = let (params, rest') = parseParams rest in ((VerilogModuleItem name) : params, rest')
parseParams (IDENTIFIER name, RPAREN : rest) = ((VerilogModuleItem name), rest)
parseParams (RPAREN : rest) = ([], rest)
parseParams _ = error "parseParams"

parseInterior

-- returns (module name, rest of tokens)
parseModule :: [Token] -> (VerilogModule, [Token])
parseModule [] = error "empty module"
parseModule (MODULE : IDENTIFIER name : rest) = let (params, rest') = parseParams rest in (VerilogModule name params [], rest')