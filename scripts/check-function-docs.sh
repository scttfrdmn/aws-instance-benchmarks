#!/bin/bash
# check-function-docs.sh
# Ensures all exported functions have comprehensive documentation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Documentation requirements
MIN_COMMENT_LINES=2  # Minimum lines of documentation for complex functions
REQUIRED_SECTIONS=("Purpose" "Parameters" "Returns" "Example")

echo "🔍 Checking function documentation..."

# Find all exported functions without proper documentation
undocumented_functions=()
poorly_documented=()

# Process each Go file
for file in $(find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go"); do
    echo "Checking $file..."
    
    # Extract exported functions
    exported_funcs=$(grep -n "^func [A-Z]" "$file" || true)
    
    while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            line_num=$(echo "$line" | cut -d: -f1)
            func_signature=$(echo "$line" | cut -d: -f2-)
            func_name=$(echo "$func_signature" | sed 's/^func \([A-Za-z0-9_]*\).*/\1/')
            
            # Check if function has documentation
            comment_start=$((line_num - 1))
            
            # Look for comment block above function
            comment_lines=""
            for ((i=comment_start; i>=1; i--)); do
                check_line=$(sed -n "${i}p" "$file")
                if [[ "$check_line" =~ ^[[:space:]]*// ]]; then
                    comment_lines="$check_line\n$comment_lines"
                elif [[ "$check_line" =~ ^[[:space:]]*$ ]]; then
                    continue  # Skip empty lines
                else
                    break  # Hit non-comment, non-empty line
                fi
            done
            
            # Validate documentation
            if [[ -z "$comment_lines" ]]; then
                undocumented_functions+=("$file:$line_num:$func_name")
            else
                # Check for comprehensive documentation
                comment_line_count=$(echo -e "$comment_lines" | wc -l)
                
                # For complex functions, require detailed documentation
                func_complexity=$(echo "$func_signature" | grep -o "(" | wc -l)
                if [[ $func_complexity -gt 2 ]] && [[ $comment_line_count -lt $MIN_COMMENT_LINES ]]; then
                    poorly_documented+=("$file:$line_num:$func_name (complex function needs detailed docs)")
                fi
                
                # Check if function name is repeated in comment (good practice)
                if ! echo -e "$comment_lines" | grep -q "$func_name"; then
                    poorly_documented+=("$file:$line_num:$func_name (comment should mention function name)")
                fi
            fi
        fi
    done <<< "$exported_funcs"
done

# Check for package-level documentation
echo "📦 Checking package documentation..."
missing_package_docs=()

for file in $(find . -name "*.go" -not -path "./vendor/*" -not -name "*_test.go"); do
    # Check if this is a package declaration file
    if grep -q "^package " "$file"; then
        package_name=$(grep "^package " "$file" | head -1 | awk '{print $2}')
        
        # Skip main package
        if [[ "$package_name" == "main" ]]; then
            continue
        fi
        
        # Check for package documentation
        if ! grep -q "^// Package $package_name" "$file" && ! grep -q "^// $package_name" "$file"; then
            missing_package_docs+=("$file:$package_name")
        fi
    fi
done

# Report results
echo ""
echo "📊 Documentation Analysis Results:"
echo "=================================="

if [[ ${#undocumented_functions[@]} -eq 0 ]] && [[ ${#poorly_documented[@]} -eq 0 ]] && [[ ${#missing_package_docs[@]} -eq 0 ]]; then
    echo -e "${GREEN}✅ All functions and packages are properly documented!${NC}"
    exit 0
fi

# Report undocumented functions
if [[ ${#undocumented_functions[@]} -gt 0 ]]; then
    echo -e "${RED}❌ Undocumented exported functions:${NC}"
    for func in "${undocumented_functions[@]}"; do
        echo "  - $func"
    done
    echo ""
fi

# Report poorly documented functions
if [[ ${#poorly_documented[@]} -gt 0 ]]; then
    echo -e "${YELLOW}⚠️  Functions needing better documentation:${NC}"
    for func in "${poorly_documented[@]}"; do
        echo "  - $func"
    done
    echo ""
fi

# Report missing package docs
if [[ ${#missing_package_docs[@]} -gt 0 ]]; then
    echo -e "${RED}📦 Missing package documentation:${NC}"
    for pkg in "${missing_package_docs[@]}"; do
        echo "  - $pkg"
    done
    echo ""
fi

echo -e "${YELLOW}Documentation Standards:${NC}"
echo "• All exported functions must have documentation comments"
echo "• Complex functions (3+ parameters) need detailed explanations"
echo "• Comments should mention the function name"
echo "• All packages must have package-level documentation"
echo "• Use format: // FunctionName does X, Y, and Z."
echo ""

exit 1