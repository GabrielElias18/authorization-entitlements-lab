#!/bin/bash
set -e

echo "📊 Loading relationships from tuples.csv..."

# Check if tuples.csv exists
if [ ! -f tuples.csv ]; then
    echo "❌ tuples.csv not found!"
    exit 1
fi

# Function to create a relationship with proper error handling
create_relationship() {
    local resource="$1"
    local relation="$2"
    local subject="$3"
    local caveat="$4"
    
    echo "Creating: $resource $relation $subject${caveat:+ with caveat $caveat}"
    
    if [ -n "$caveat" ]; then
        zed relationship create "$resource" "$relation" "$subject" --caveat "$caveat"
    else
        zed relationship create "$resource" "$relation" "$subject"
    fi
    
    if [ $? -eq 0 ]; then
        echo "✅ Successfully created relationship"
    else
        echo "❌ Failed to create relationship"
        return 1
    fi
}

# Skip header and process each line
line_number=1
while IFS= read -r line; do
    line_number=$((line_number + 1))
    
    # Skip empty lines
    if [ -z "$line" ]; then
        continue
    fi
    
    # Parse CSV line (handle commas within JSON)
    # Use a simple approach: split on first 3 commas, rest is caveat
    resource=$(echo "$line" | cut -d',' -f1 | tr -d ' ')
    relation=$(echo "$line" | cut -d',' -f2 | tr -d ' ')
    subject=$(echo "$line" | cut -d',' -f3 | tr -d ' ')
    caveat=$(echo "$line" | cut -d',' -f4- | tr -d ' ')
    
    # Validate required fields
    if [ -z "$resource" ] || [ -z "$relation" ] || [ -z "$subject" ]; then
        echo "⚠️  Skipping line $line_number: missing required fields"
        echo "   Line: $line"
        continue
    fi
    
    # Handle empty caveat (remove trailing comma)
    if [ "$caveat" = "" ]; then
        caveat=""
    fi
    
    # Create the relationship
    if ! create_relationship "$resource" "$relation" "$subject" "$caveat"; then
        echo "❌ Failed to create relationship on line $line_number"
        echo "   Line: $line"
        exit 1
    fi
    
done < <(tail -n +2 tuples.csv)

echo "✅ All relationships loaded successfully!"
echo ""
echo "📋 Summary of loaded relationships:"
zed relationship read account
zed relationship read poa
zed relationship read role
zed relationship read org 